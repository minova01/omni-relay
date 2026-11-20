package cataloghttp

import (
	"errors"
	"log/slog"
	"net/http"

	"omni-relay/internal/catalog/application"
	"omni-relay/internal/catalog/domain"
	httpmodel "omni-relay/internal/catalog/interfaces/http/model"
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

func NewHandler(service *application.Service, identity *identityapp.Service, logger *slog.Logger) *Handler {
	return &Handler{service: service, identity: identity, logger: logger}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.Handle("POST /api/v1/admin/channels", h.requireAdmin(http.HandlerFunc(h.createChannel)))
	mux.Handle("GET /api/v1/admin/channels", h.requireAdmin(http.HandlerFunc(h.listChannels)))
	mux.Handle("POST /api/v1/admin/channels/update", h.requireAdmin(http.HandlerFunc(h.updateChannel)))
	mux.Handle("POST /api/v1/admin/channels/status", h.requireAdmin(http.HandlerFunc(h.updateChannelStatus)))

	mux.Handle("POST /api/v1/admin/models", h.requireAdmin(http.HandlerFunc(h.createModel)))
	mux.Handle("GET /api/v1/admin/models", h.requireAdmin(http.HandlerFunc(h.listModels)))
	mux.Handle("POST /api/v1/admin/models/update", h.requireAdmin(http.HandlerFunc(h.updateModel)))
	mux.Handle("POST /api/v1/admin/models/status", h.requireAdmin(http.HandlerFunc(h.updateModelStatus)))

	mux.Handle("POST /api/v1/admin/channel-models", h.requireAdmin(http.HandlerFunc(h.createChannelModel)))
	mux.Handle("GET /api/v1/admin/channel-models", h.requireAdmin(http.HandlerFunc(h.listChannelModels)))
	mux.Handle("POST /api/v1/admin/channel-models/update", h.requireAdmin(http.HandlerFunc(h.updateChannelModel)))
	mux.Handle("POST /api/v1/admin/channel-models/status", h.requireAdmin(http.HandlerFunc(h.updateChannelModelStatus)))
	mux.Handle("POST /api/v1/admin/channel-models/delete", h.requireAdmin(http.HandlerFunc(h.deleteChannelModel)))
}

func (h *Handler) createChannel(w http.ResponseWriter, r *http.Request) {
	var request httpmodel.CreateChannelRequest
	if !decodeRequest(w, r, &request) {
		return
	}
	channel, err := h.service.CreateChannel(r.Context(), application.ChannelInput{
		Name: request.Name, ChannelType: request.ChannelType, BaseURL: request.BaseURL, Secret: request.Secret,
	})
	if err != nil {
		h.writeError(w, r, "create channel", err)
		return
	}
	httpserver.WriteJSON(w, http.StatusCreated, map[string]any{"data": httpmodel.NewChannelResponse(channel)})
}

func (h *Handler) listChannels(w http.ResponseWriter, r *http.Request) {
	channels, err := h.service.ListChannels(r.Context())
	if err != nil {
		h.writeError(w, r, "list channels", err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, map[string]any{"data": httpmodel.NewChannelResponses(channels)})
}

func (h *Handler) updateChannel(w http.ResponseWriter, r *http.Request) {
	var request httpmodel.UpdateChannelRequest
	if !decodeRequest(w, r, &request) {
		return
	}
	err := h.service.UpdateChannel(r.Context(), request.ID, application.ChannelInput{
		Name: request.Name, ChannelType: request.ChannelType, BaseURL: request.BaseURL, Secret: request.Secret,
	})
	if err != nil {
		h.writeError(w, r, "update channel", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) updateChannelStatus(w http.ResponseWriter, r *http.Request) {
	var request httpmodel.StatusRequest
	if !decodeRequest(w, r, &request) {
		return
	}
	if err := h.service.UpdateChannelStatus(r.Context(), request.ID, domain.Status(request.Status)); err != nil {
		h.writeError(w, r, "update channel status", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) createModel(w http.ResponseWriter, r *http.Request) {
	var request httpmodel.CreateModelRequest
	if !decodeRequest(w, r, &request) {
		return
	}
	model, err := h.service.CreateModel(r.Context(), application.ModelInput{
		Code: request.Code, DisplayName: request.DisplayName,
		InputPrice: request.InputPrice, OutputPrice: request.OutputPrice,
	})
	if err != nil {
		h.writeError(w, r, "create model", err)
		return
	}
	httpserver.WriteJSON(w, http.StatusCreated, map[string]any{"data": httpmodel.NewModelResponse(model)})
}

func (h *Handler) listModels(w http.ResponseWriter, r *http.Request) {
	models, err := h.service.ListModels(r.Context())
	if err != nil {
		h.writeError(w, r, "list models", err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, map[string]any{"data": httpmodel.NewModelResponses(models)})
}

func (h *Handler) updateModel(w http.ResponseWriter, r *http.Request) {
	var request httpmodel.UpdateModelRequest
	if !decodeRequest(w, r, &request) {
		return
	}
	err := h.service.UpdateModel(r.Context(), request.ID, application.ModelInput{
		Code: request.Code, DisplayName: request.DisplayName,
		InputPrice: request.InputPrice, OutputPrice: request.OutputPrice,
	})
	if err != nil {
		h.writeError(w, r, "update model", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) updateModelStatus(w http.ResponseWriter, r *http.Request) {
	var request httpmodel.StatusRequest
	if !decodeRequest(w, r, &request) {
		return
	}
	if err := h.service.UpdateModelStatus(r.Context(), request.ID, domain.Status(request.Status)); err != nil {
		h.writeError(w, r, "update model status", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) createChannelModel(w http.ResponseWriter, r *http.Request) {
	var request httpmodel.CreateChannelModelRequest
	if !decodeRequest(w, r, &request) {
		return
	}
	mapping, err := h.service.CreateChannelModel(r.Context(), channelModelInput(request.ChannelID, request.ModelID, request.UpstreamModel, request.Priority, request.SupportsStream, request.SupportsTools))
	if err != nil {
		h.writeError(w, r, "create channel model", err)
		return
	}
	httpserver.WriteJSON(w, http.StatusCreated, map[string]any{"data": httpmodel.NewChannelModelResponse(mapping)})
}

func (h *Handler) listChannelModels(w http.ResponseWriter, r *http.Request) {
	mappings, err := h.service.ListChannelModels(r.Context())
	if err != nil {
		h.writeError(w, r, "list channel models", err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, map[string]any{"data": httpmodel.NewChannelModelResponses(mappings)})
}

func (h *Handler) updateChannelModel(w http.ResponseWriter, r *http.Request) {
	var request httpmodel.UpdateChannelModelRequest
	if !decodeRequest(w, r, &request) {
		return
	}
	err := h.service.UpdateChannelModel(r.Context(), request.ID, channelModelInput(request.ChannelID, request.ModelID, request.UpstreamModel, request.Priority, request.SupportsStream, request.SupportsTools))
	if err != nil {
		h.writeError(w, r, "update channel model", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) updateChannelModelStatus(w http.ResponseWriter, r *http.Request) {
	var request httpmodel.StatusRequest
	if !decodeRequest(w, r, &request) {
		return
	}
	if err := h.service.UpdateChannelModelStatus(r.Context(), request.ID, domain.Status(request.Status)); err != nil {
		h.writeError(w, r, "update channel model status", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) deleteChannelModel(w http.ResponseWriter, r *http.Request) {
	var request httpmodel.DeleteChannelModelRequest
	if !decodeRequest(w, r, &request) {
		return
	}
	if err := h.service.DeleteChannelModel(r.Context(), request.ID); err != nil {
		h.writeError(w, r, "delete channel model", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) requireAdmin(next http.Handler) http.Handler {
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
				h.logError(r, "authenticate admin session", err)
				httpserver.WriteError(w, http.StatusInternalServerError, "internal_error", "服务暂时不可用")
			}
			return
		}
		if user.Role != identitydomain.RoleAdmin {
			httpserver.WriteError(w, http.StatusForbidden, "admin_required", "需要管理员权限")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (h *Handler) writeError(w http.ResponseWriter, r *http.Request, operation string, err error) {
	switch {
	case errors.Is(err, domain.ErrInvalidBaseURL):
		httpserver.WriteError(w, http.StatusBadRequest, "invalid_base_url", "渠道地址不正确")
	case errors.Is(err, domain.ErrChannelNameTaken):
		httpserver.WriteError(w, http.StatusConflict, "channel_name_taken", "渠道名称已存在")
	case errors.Is(err, domain.ErrModelCodeTaken):
		httpserver.WriteError(w, http.StatusConflict, "model_code_taken", "模型代码已存在")
	case errors.Is(err, domain.ErrChannelModelTaken):
		httpserver.WriteError(w, http.StatusConflict, "channel_model_taken", "渠道模型映射已存在")
	case errors.Is(err, domain.ErrModelHasNoActiveRoute):
		httpserver.WriteError(w, http.StatusConflict, "model_has_no_active_route", "模型没有启用的渠道映射")
	case errors.Is(err, domain.ErrChannelNotFound):
		httpserver.WriteError(w, http.StatusNotFound, "channel_not_found", "渠道不存在")
	case errors.Is(err, domain.ErrModelNotFound):
		httpserver.WriteError(w, http.StatusNotFound, "model_not_found", "模型不存在")
	case errors.Is(err, domain.ErrChannelModelNotFound):
		httpserver.WriteError(w, http.StatusNotFound, "channel_model_not_found", "渠道模型映射不存在")
	default:
		h.logError(r, operation, err)
		httpserver.WriteError(w, http.StatusInternalServerError, "internal_error", "服务暂时不可用")
	}
}

func (h *Handler) logError(r *http.Request, operation string, err error) {
	h.logger.ErrorContext(r.Context(), operation,
		"request_id", httpserver.RequestID(r.Context()), "error", err,
	)
}

func decodeRequest(w http.ResponseWriter, r *http.Request, destination any) bool {
	if err := httpserver.DecodeJSON(w, r, destination, maxRequestBodyBytes); err != nil {
		httpserver.WriteError(w, http.StatusBadRequest, "invalid_request", "请求参数不正确")
		return false
	}
	return true
}

func channelModelInput(channelID, modelID int64, upstreamModel string, priority int, supportsStream, supportsTools bool) application.ChannelModelInput {
	return application.ChannelModelInput{
		ChannelID: channelID, ModelID: modelID, UpstreamModel: upstreamModel,
		Priority: priority, SupportsStream: supportsStream, SupportsTools: supportsTools,
	}
}
