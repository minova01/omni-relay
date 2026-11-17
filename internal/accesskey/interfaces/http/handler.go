package accesskeyhttp

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"omni-relay/internal/accesskey/application"
	"omni-relay/internal/accesskey/domain"
	httpmodel "omni-relay/internal/accesskey/interfaces/http/model"
	identityapp "omni-relay/internal/identity/application"
	identitydomain "omni-relay/internal/identity/domain"
	identityhttp "omni-relay/internal/identity/interfaces/http"
	"omni-relay/internal/shared/infrastructure/httpserver"
)

const maxRequestBodyBytes = 1 << 20

type Handler struct {
	service  *application.Service
	identity *identityapp.Service
	logger   *slog.Logger
}

type sessionUserContextKey struct{}

func NewHandler(service *application.Service, identity *identityapp.Service, logger *slog.Logger) *Handler {
	return &Handler{service: service, identity: identity, logger: logger}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.Handle("POST /api/v1/", h.requireSession(http.HandlerFunc(h.create)))
	mux.Handle("GET /api/v1/api-keys", h.requireSession(http.HandlerFunc(h.list)))
	mux.Handle("POST /api/v1/api-keys/{id}/status", h.requireSession(http.HandlerFunc(h.updateStatus)))
	mux.Handle("POST /api/v1/api-keys/{id}/delete", h.requireSession(http.HandlerFunc(h.delete)))
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var request httpmodel.CreateAPIKeyRequest
	if err := httpserver.DecodeJSON(w, r, &request, maxRequestBodyBytes); err != nil {
		httpserver.WriteError(w, http.StatusBadRequest, "invalid_request", "请求参数不正确")
		return
	}
	user := sessionUser(r.Context())
	result, err := h.service.Create(r.Context(), user.ID, request.Name)
	if err != nil {
		h.logError(r, "create api key", err)
		httpserver.WriteError(w, http.StatusInternalServerError, "internal_error", "服务暂时不可用")
		return
	}
	httpserver.WriteJSON(w, http.StatusCreated, map[string]any{
		"data": httpmodel.NewCreatedAPIKeyResponse(result.APIKey, result.Plaintext),
	})
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	user := sessionUser(r.Context())
	keys, err := h.service.List(r.Context(), user.ID)
	if err != nil {
		h.logError(r, "list api keys", err)
		httpserver.WriteError(w, http.StatusInternalServerError, "internal_error", "服务暂时不可用")
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, map[string]any{"data": httpmodel.NewAPIKeyResponses(keys)})
}

func (h *Handler) updateStatus(w http.ResponseWriter, r *http.Request) {
	keyID, ok := pathID(w, r)
	if !ok {
		return
	}
	var request httpmodel.UpdateAPIKeyStatusRequest
	if err := httpserver.DecodeJSON(w, r, &request, maxRequestBodyBytes); err != nil {
		httpserver.WriteError(w, http.StatusBadRequest, "invalid_request", "请求参数不正确")
		return
	}
	user := sessionUser(r.Context())
	if err := h.service.UpdateStatus(r.Context(), user.ID, keyID, domain.Status(request.Status)); err != nil {
		h.writeMutationError(w, r, "update api key status", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	keyID, ok := pathID(w, r)
	if !ok {
		return
	}
	user := sessionUser(r.Context())
	if err := h.service.Delete(r.Context(), user.ID, keyID); err != nil {
		h.writeMutationError(w, r, "delete api key", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) requireSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(identityhttp.SessionCookieName)
		if err != nil {
			httpserver.WriteError(w, http.StatusUnauthorized, "invalid_session", "请先登录")
			return
		}
		user, err := h.identity.AuthenticateSession(r.Context(), cookie.Value)
		if err != nil {
			switch {
			case errors.Is(err, identitydomain.ErrInvalidSession):
				httpserver.WriteError(w, http.StatusUnauthorized, "invalid_session", "请先登录")
			case errors.Is(err, identitydomain.ErrUserDisabled):
				httpserver.WriteError(w, http.StatusForbidden, "user_disabled", "用户已被禁用")
			default:
				h.logError(r, "authenticate session", err)
				httpserver.WriteError(w, http.StatusInternalServerError, "internal_error", "服务暂时不可用")
			}
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), sessionUserContextKey{}, user)))
	})
}

func (h *Handler) RequireAPIKey(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		plaintext, ok := bearerToken(r.Header.Get("Authorization"))
		if !ok {
			httpserver.WriteError(w, http.StatusUnauthorized, "invalid_api_key", "API Key 无效")
			return
		}
		principal, err := h.service.Authenticate(r.Context(), plaintext)
		if err != nil {
			switch {
			case errors.Is(err, domain.ErrInvalidAPIKey):
				httpserver.WriteError(w, http.StatusUnauthorized, "invalid_api_key", "API Key 无效")
			case errors.Is(err, domain.ErrAPIKeyDisabled), errors.Is(err, identitydomain.ErrUserDisabled):
				httpserver.WriteError(w, http.StatusForbidden, "access_disabled", "API Key 或用户已被禁用")
			case errors.Is(err, domain.ErrInsufficientQuota):
				httpserver.WriteError(w, http.StatusPaymentRequired, "insufficient_quota", "用户额度不足")
			default:
				h.logError(r, "authenticate api key", err)
				httpserver.WriteError(w, http.StatusInternalServerError, "internal_error", "服务暂时不可用")
			}
			return
		}
		next.ServeHTTP(w, r.WithContext(application.WithPrincipal(r.Context(), principal)))
	})
}

func (h *Handler) writeMutationError(w http.ResponseWriter, r *http.Request, operation string, err error) {
	if errors.Is(err, domain.ErrAPIKeyNotFound) {
		httpserver.WriteError(w, http.StatusNotFound, "api_key_not_found", "API Key 不存在")
		return
	}
	h.logError(r, operation, err)
	httpserver.WriteError(w, http.StatusInternalServerError, "internal_error", "服务暂时不可用")
}

func (h *Handler) logError(r *http.Request, operation string, err error) {
	h.logger.ErrorContext(r.Context(), operation,
		"request_id", httpserver.RequestID(r.Context()),
		"error", err,
	)
}

func pathID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		httpserver.WriteError(w, http.StatusBadRequest, "invalid_id", "API Key ID 不正确")
		return 0, false
	}
	return id, true
}

func bearerToken(header string) (string, bool) {
	parts := strings.Fields(header)
	returnToken := ""
	if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
		returnToken = parts[1]
	}
	return returnToken, returnToken != ""
}

func sessionUser(ctx context.Context) identitydomain.User {
	return ctx.Value(sessionUserContextKey{}).(identitydomain.User)
}
