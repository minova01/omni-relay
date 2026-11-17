package application

import (
	"context"
	"time"
)

type PasswordService interface {
	Hash(string) (string, error)
	Verify(string, string) bool
	DummyVerify(string)
}

type SessionStore interface {
	Create(context.Context, int64, time.Duration) (string, error)
	Delete(context.Context, string) error
}
