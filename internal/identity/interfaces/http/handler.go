package identityhttp

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"

	"omni-relay/internal/identity/application"
	"omni-relay/internal/identity/domain"
	httpmodel "omni-relay/internal/identity/interfaces/http/model"
	"omni-relay/internal/shared/infrastructure/httpserver"
)

const (
	maxRequestBodyBytes = 1 << 20
	sessionCookieName   = "omni_session"
)

type Handler struct {
	service      *application.Service
	logger       *slog.Logger
	sessionTTL   time.Duration
	cookieSecure bool
}

func NewHandler(
	service *application.Service,
	logger *slog.Logger,
	sessionTTL time.Duration,
	cookieSecure bool,
) *Handler {
	return &Handler{
		service:      service,
		logger:       logger,
		sessionTTL:   sessionTTL,
		cookieSecure: cookieSecure,
	}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/auth/register", h.register)
	mux.HandleFunc("POST /api/v1/auth/login", h.login)
	mux.HandleFunc("POST /api/v1/auth/logout", h.logout)
}

func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	var request httpmodel.CredentialsRequest
	if err := httpserver.DecodeJSON(w, r, &request, maxRequestBodyBytes); err != nil {
		switch invalidField(err) {
		case "Username":
			writeError(w, http.StatusBadRequest, "invalid_username", "用户名格式不正确")
		case "Password":
			writeError(w, http.StatusBadRequest, "invalid_password", "密码长度必须为 8 到 72 字节")
		default:
			writeError(w, http.StatusBadRequest, "invalid_request", "请求参数不正确")
		}
		return
	}

	user, err := h.service.Register(r.Context(), request.Username, request.Password)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrUsernameTaken):
			writeError(w, http.StatusConflict, "username_taken", "用户名已存在")
		default:
			h.logError(r, "register user", err)
			writeError(w, http.StatusInternalServerError, "internal_error", "服务暂时不可用")
		}
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": httpmodel.NewUserResponse(user)})
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var request httpmodel.CredentialsRequest
	if err := httpserver.DecodeJSON(w, r, &request, maxRequestBodyBytes); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "请求参数不正确")
		return
	}

	result, err := h.service.Login(r.Context(), request.Username, request.Password)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidCredential):
			writeError(w, http.StatusUnauthorized, "invalid_credentials", "用户名或密码错误")
		case errors.Is(err, domain.ErrUserDisabled):
			writeError(w, http.StatusForbidden, "user_disabled", "用户已被禁用")
		default:
			h.logError(r, "login user", err)
			writeError(w, http.StatusInternalServerError, "internal_error", "服务暂时不可用")
		}
		return
	}

	h.setSessionCookie(w, result.SessionToken)
	writeJSON(w, http.StatusOK, map[string]any{"data": httpmodel.NewUserResponse(result.User)})
}

func invalidField(err error) string {
	var validationErrors validator.ValidationErrors
	if errors.As(err, &validationErrors) && len(validationErrors) > 0 {
		return validationErrors[0].StructField()
	}
	return ""
}

func (h *Handler) logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil && !errors.Is(err, http.ErrNoCookie) {
		writeError(w, http.StatusBadRequest, "invalid_session", "会话信息不正确")
		return
	}
	if cookie != nil {
		if err := h.service.Logout(r.Context(), cookie.Value); err != nil {
			h.logError(r, "logout user", err)
			writeError(w, http.StatusInternalServerError, "internal_error", "服务暂时不可用")
			return
		}
	}
	h.clearSessionCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) setSessionCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   int(h.sessionTTL.Seconds()),
		HttpOnly: true,
		Secure:   h.cookieSecure,
		SameSite: http.SameSiteLaxMode,
	})
}

func (h *Handler) clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   h.cookieSecure,
		SameSite: http.SameSiteLaxMode,
	})
}

func (h *Handler) logError(r *http.Request, operation string, err error) {
	h.logger.ErrorContext(r.Context(), operation,
		"request_id", httpserver.RequestID(r.Context()),
		"error", err,
	)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]any{
		"error": map[string]string{"code": code, "message": message},
	})
}
