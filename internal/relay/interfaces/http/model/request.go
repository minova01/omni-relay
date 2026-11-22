package model

import "encoding/json"

type Message struct {
	Role    string `json:"role" validate:"required,oneof=system developer user assistant tool"`
	Content string `json:"content" validate:"required"`
}

type ChatRequest struct {
	Model       string          `json:"model" validate:"required"`
	Messages    []Message       `json:"messages" validate:"required,min=1,dive"`
	Stream      bool            `json:"stream"`
	MaxTokens   *int            `json:"max_tokens,omitempty" validate:"omitempty,gt=0"`
	Temperature *float64        `json:"temperature,omitempty" validate:"omitempty,gte=0,lte=2"`
	TopP        *float64        `json:"top_p,omitempty" validate:"omitempty,gt=0,lte=1"`
	Stop        json.RawMessage `json:"stop,omitempty"`
}
