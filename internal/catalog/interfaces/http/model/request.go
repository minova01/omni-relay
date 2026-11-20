package model

type CreateChannelRequest struct {
	Name        string `json:"name" validate:"required,notblank,max=128"`
	ChannelType string `json:"channel_type" validate:"required,oneof=openai_compatible"`
	BaseURL     string `json:"base_url" validate:"required,url,max=512"`
	Secret      string `json:"secret" validate:"required,notblank,max=4096"`
}

type UpdateChannelRequest struct {
	ID          int64  `json:"id" validate:"required,gt=0"`
	Name        string `json:"name" validate:"required,notblank,max=128"`
	ChannelType string `json:"channel_type" validate:"required,oneof=openai_compatible"`
	BaseURL     string `json:"base_url" validate:"required,url,max=512"`
	Secret      string `json:"secret,omitempty" validate:"omitempty,notblank,max=4096"`
}

type StatusRequest struct {
	ID     int64 `json:"id" validate:"required,gt=0"`
	Status int16 `json:"status" validate:"required,oneof=1 2"`
}

type CreateModelRequest struct {
	Code        string `json:"code" validate:"required,notblank,max=128"`
	DisplayName string `json:"display_name" validate:"required,notblank,max=128"`
	InputPrice  int64  `json:"input_price" validate:"gte=0"`
	OutputPrice int64  `json:"output_price" validate:"gte=0"`
}

type UpdateModelRequest struct {
	ID          int64  `json:"id" validate:"required,gt=0"`
	Code        string `json:"code" validate:"required,notblank,max=128"`
	DisplayName string `json:"display_name" validate:"required,notblank,max=128"`
	InputPrice  int64  `json:"input_price" validate:"gte=0"`
	OutputPrice int64  `json:"output_price" validate:"gte=0"`
}

type CreateChannelModelRequest struct {
	ChannelID      int64  `json:"channel_id" validate:"required,gt=0"`
	ModelID        int64  `json:"model_id" validate:"required,gt=0"`
	UpstreamModel  string `json:"upstream_model" validate:"required,notblank,max=128"`
	Priority       int    `json:"priority" validate:"gte=-2147483648,lte=2147483647"`
	SupportsStream bool   `json:"supports_stream"`
	SupportsTools  bool   `json:"supports_tools"`
}

type UpdateChannelModelRequest struct {
	ID             int64  `json:"id" validate:"required,gt=0"`
	ChannelID      int64  `json:"channel_id" validate:"required,gt=0"`
	ModelID        int64  `json:"model_id" validate:"required,gt=0"`
	UpstreamModel  string `json:"upstream_model" validate:"required,notblank,max=128"`
	Priority       int    `json:"priority" validate:"gte=-2147483648,lte=2147483647"`
	SupportsStream bool   `json:"supports_stream"`
	SupportsTools  bool   `json:"supports_tools"`
}

type DeleteChannelModelRequest struct {
	ID int64 `json:"id" validate:"required,gt=0"`
}
