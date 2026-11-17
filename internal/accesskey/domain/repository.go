package domain

import "context"

type Repository interface {
	Create(context.Context, *APIKey) error
	ListByUser(context.Context, int64) ([]APIKey, error)
	FindByPrefix(context.Context, string) (APIKey, error)
	UpdateStatus(context.Context, int64, int64, Status, int64) error
	SoftDelete(context.Context, int64, int64, int64) error
}
