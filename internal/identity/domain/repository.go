package domain

import "context"

type UserRepository interface {
	Create(context.Context, *User) error
	FindByID(context.Context, int64) (User, error)
	FindByUsername(context.Context, string) (User, error)
}
