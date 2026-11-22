package relayhttp

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	accesskeyhttp "omni-relay/internal/accesskey/interfaces/http"
	"omni-relay/internal/relay/infrastructure/postgres"
	"omni-relay/internal/relay/interfaces/http/model"
	"omni-relay/internal/shared/infrastructure/httpserver"
	"omni-relay/internal/shared/infrastructure/secret"
)

type Handler struct {
	routes  *postgres.Repository
	cipher  *secret.Cipher
	client  *http.Client
	maxBody int64
	logger  *slog.Logger
}

func NewHandler(routes *postgres.Repository, cipher *secret.Cipher, timeout time.Duration, maxBody int64, logger *slog.Logger) *Handler {
	return &Handler{routes: routes, cipher: cipher, client: &http.Client{Timeout: timeout}, maxBody: maxBody, logger: logger}
}

// register the routes for the relay handler
func (h *Handler) RegisterRoutes(mux *http.ServeMux, keys *accesskeyhttp.Handler) {
	mux.Handle("POST /v1/chat/completions", keys.RequireAPIKey(http.HandlerFunc(h.chat)))
}
func (h *Handler) chat(w http.ResponseWriter, r *http.Request) {
	var input model.ChatRequest
	if err := httpserver.DecodeJSON(w, r, &input, h.maxBody); err != nil {
		httpserver.WriteError(w, http.StatusBadRequest, "invalid_request", "请求参数不正确")
		return
	}
	if !validStop(input.Stop) {
		httpserver.WriteError(w, http.StatusBadRequest, "invalid_request", "stop 参数不正确")
		return
	}
	route, err := h.routes.Select(r.Context(), input.Model)
	if err != nil {
		if errors.Is(err, postgres.ErrModelUnavailable) {
			httpserver.WriteError(w, http.StatusNotFound, "model_not_found", "模型不存在或未开放")
		} else {
			h.fail(w, r, err)
		}
		return
	}
	if route.Channel.ChannelType != "openai_compatible" {
		h.fail(w, r, errors.New("unsupported channel type"))
		return
	}
	if input.Stream && !route.Mapping.SupportsStream {
		httpserver.WriteError(w, http.StatusBadRequest, "stream_unsupported", "当前模型渠道不支持流式响应")
		return
	}
	credential, err := h.cipher.Decrypt(route.Channel.SecretCiphertext)
	if err != nil {
		h.fail(w, r, err)
		return
	}
	input.Model = route.Mapping.UpstreamModel
	if input.Stream {
		input.StreamOptions = &model.StreamOptions{IncludeUsage: true}
	}
	body, err := json.Marshal(input)
	if err != nil {
		h.fail(w, r, err)
		return
	}
	url := strings.TrimRight(route.Channel.BaseURL, "/") + "/chat/completions"
	upstream, err := http.NewRequestWithContext(r.Context(), http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		h.fail(w, r, err)
		return
	}
	upstream.Header.Set("Authorization", "Bearer "+credential)
	upstream.Header.Set("Content-Type", "application/json")
	response, err := h.client.Do(upstream)
	if err != nil {
		if errors.Is(err, r.Context().Err()) {
			return
		}
		if timeout, ok := err.(interface{ Timeout() bool }); ok && timeout.Timeout() {
			httpserver.WriteError(w, http.StatusGatewayTimeout, "upstream_timeout", "上游请求超时")
			return
		}
		h.fail(w, r, err)
		return
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		status := http.StatusBadGateway
		if response.StatusCode == 429 {
			status = http.StatusTooManyRequests
		}
		httpserver.WriteError(w, status, "upstream_error", "上游服务调用失败")
		return
	}
	if input.Stream {
		h.stream(w, r, response)
		return
	}
	limited := io.LimitReader(response.Body, h.maxBody+1)
	data, err := io.ReadAll(limited)
	if err != nil || int64(len(data)) > h.maxBody || !json.Valid(data) {
		h.fail(w, r, errors.New("invalid upstream response"))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}
func validStop(raw json.RawMessage) bool {
	if len(raw) == 0 {
		return true
	}
	var value any
	if json.Unmarshal(raw, &value) != nil {
		return false
	}
	switch v := value.(type) {
	case string:
		return v != ""
	case []any:
		if len(v) == 0 {
			return false
		}
		for _, item := range v {
			s, ok := item.(string)
			if !ok || s == "" {
				return false
			}
		}
		return true
	}
	return false
}
func (h *Handler) fail(w http.ResponseWriter, r *http.Request, err error) {
	h.logger.ErrorContext(r.Context(), "relay request failed", "request_id", httpserver.RequestID(r.Context()), "error", err)
	httpserver.WriteError(w, http.StatusBadGateway, "upstream_error", "上游服务调用失败")
}
