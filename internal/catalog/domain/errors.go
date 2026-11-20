package domain

import "errors"

var (
	ErrChannelNotFound       = errors.New("channel not found")
	ErrChannelNameTaken      = errors.New("channel name already exists")
	ErrModelNotFound         = errors.New("model not found")
	ErrModelCodeTaken        = errors.New("model code already exists")
	ErrChannelModelNotFound  = errors.New("channel model not found")
	ErrChannelModelTaken     = errors.New("channel model already exists")
	ErrModelHasNoActiveRoute = errors.New("model has no active route")
	ErrInvalidBaseURL        = errors.New("invalid channel base url")
)
