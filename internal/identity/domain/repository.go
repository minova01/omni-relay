package domain

import "context"

type UserRepository interface {
	Create(context.Context, *User) error
	FindByUsername(context.Context, string) (User, error)
}
