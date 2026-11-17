package domain

import "errors"

var (
	ErrAPIKeyNotFound    = errors.New("api key not found")
	ErrAPIKeyDisabled    = errors.New("api key is disabled")
	ErrInvalidAPIKey     = errors.New("invalid api key")
	ErrPrefixConflict    = errors.New("api key prefix conflict")
	ErrInsufficientQuota = errors.New("insufficient quota")
)
