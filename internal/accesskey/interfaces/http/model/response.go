package model

import "omni-relay/internal/accesskey/domain"

type APIKeyResponse struct {
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	Prefix       string `json:"prefix"`
	Hint         string `json:"hint"`
	Status       int16  `json:"status"`
	UsedQuota    int64  `json:"used_quota"`
	RequestCount int64  `json:"request_count"`
	LastUsedAt   int64  `json:"last_used_at"`
	CreatedAt    int64  `json:"created_at"`
	UpdatedAt    int64  `json:"updated_at"`
}

type CreatedAPIKeyResponse struct {
	APIKeyResponse
	Key string `json:"key"`
}

func NewAPIKeyResponse(key domain.APIKey) APIKeyResponse {
	return APIKeyResponse{
		ID:           key.ID,
		Name:         key.Name,
		Prefix:       "or_" + key.KeyPrefix,
		Hint:         key.KeyHint,
		Status:       int16(key.Status),
		UsedQuota:    key.UsedQuota,
		RequestCount: key.RequestCount,
		LastUsedAt:   key.LastUsedAt,
		CreatedAt:    key.CreatedAt,
		UpdatedAt:    key.UpdatedAt,
	}
}

func NewCreatedAPIKeyResponse(key domain.APIKey, plaintext string) CreatedAPIKeyResponse {
	return CreatedAPIKeyResponse{APIKeyResponse: NewAPIKeyResponse(key), Key: plaintext}
}

func NewAPIKeyResponses(keys []domain.APIKey) []APIKeyResponse {
	responses := make([]APIKeyResponse, 0, len(keys))
	for _, key := range keys {
		responses = append(responses, NewAPIKeyResponse(key))
	}
	return responses
}
