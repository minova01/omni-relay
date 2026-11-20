package model

import "omni-relay/internal/catalog/domain"

type ChannelResponse struct {
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	ChannelType  string `json:"channel_type"`
	BaseURL      string `json:"base_url"`
	Status       int16  `json:"status"`
	UsedQuota    int64  `json:"used_quota"`
	RequestCount int64  `json:"request_count"`
	CreatedAt    int64  `json:"created_at"`
	UpdatedAt    int64  `json:"updated_at"`
}

type ModelResponse struct {
	ID          int64  `json:"id"`
	Code        string `json:"code"`
	DisplayName string `json:"display_name"`
	Status      int16  `json:"status"`
	InputPrice  int64  `json:"input_price"`
	OutputPrice int64  `json:"output_price"`
	CreatedAt   int64  `json:"created_at"`
	UpdatedAt   int64  `json:"updated_at"`
}

type ChannelModelResponse struct {
	ID             int64  `json:"id"`
	ChannelID      int64  `json:"channel_id"`
	ModelID        int64  `json:"model_id"`
	UpstreamModel  string `json:"upstream_model"`
	Priority       int    `json:"priority"`
	SupportsStream bool   `json:"supports_stream"`
	SupportsTools  bool   `json:"supports_tools"`
	Status         int16  `json:"status"`
	CreatedAt      int64  `json:"created_at"`
	UpdatedAt      int64  `json:"updated_at"`
}

func NewChannelResponse(channel domain.Channel) ChannelResponse {
	return ChannelResponse{
		ID: channel.ID, Name: channel.Name, ChannelType: channel.ChannelType,
		BaseURL: channel.BaseURL, Status: int16(channel.Status), UsedQuota: channel.UsedQuota,
		RequestCount: channel.RequestCount, CreatedAt: channel.CreatedAt, UpdatedAt: channel.UpdatedAt,
	}
}

func NewChannelResponses(channels []domain.Channel) []ChannelResponse {
	responses := make([]ChannelResponse, 0, len(channels))
	for _, channel := range channels {
		responses = append(responses, NewChannelResponse(channel))
	}
	return responses
}

func NewModelResponse(model domain.Model) ModelResponse {
	return ModelResponse{
		ID: model.ID, Code: model.Code, DisplayName: model.DisplayName, Status: int16(model.Status),
		InputPrice: model.InputPrice, OutputPrice: model.OutputPrice,
		CreatedAt: model.CreatedAt, UpdatedAt: model.UpdatedAt,
	}
}

func NewModelResponses(models []domain.Model) []ModelResponse {
	responses := make([]ModelResponse, 0, len(models))
	for _, model := range models {
		responses = append(responses, NewModelResponse(model))
	}
	return responses
}

func NewChannelModelResponse(mapping domain.ChannelModel) ChannelModelResponse {
	return ChannelModelResponse{
		ID: mapping.ID, ChannelID: mapping.ChannelID, ModelID: mapping.ModelID,
		UpstreamModel: mapping.UpstreamModel, Priority: mapping.Priority,
		SupportsStream: mapping.SupportsStream, SupportsTools: mapping.SupportsTools,
		Status: int16(mapping.Status), CreatedAt: mapping.CreatedAt, UpdatedAt: mapping.UpdatedAt,
	}
}

func NewChannelModelResponses(mappings []domain.ChannelModel) []ChannelModelResponse {
	responses := make([]ChannelModelResponse, 0, len(mappings))
	for _, mapping := range mappings {
		responses = append(responses, NewChannelModelResponse(mapping))
	}
	return responses
}
