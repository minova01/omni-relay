package domain

import "errors"

var (
	ErrUsernameTaken     = errors.New("username already exists")
	ErrUserNotFound      = errors.New("user not found")
	ErrInvalidCredential = errors.New("invalid username or password")
	ErrInvalidSession    = errors.New("invalid session")
	ErrUserDisabled      = errors.New("user is disabled")
)
