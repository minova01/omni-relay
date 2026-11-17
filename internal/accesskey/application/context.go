package application

import (
	"context"

	"omni-relay/internal/accesskey/domain"
	identitydomain "omni-relay/internal/identity/domain"
)

type Principal struct {
	User    identitydomain.User
	APIKey  domain.APIKey
	Balance int64
}

type principalContextKey struct{}

func WithPrincipal(ctx context.Context, principal Principal) context.Context {
	return context.WithValue(ctx, principalContextKey{}, principal)
}

func PrincipalFromContext(ctx context.Context) (Principal, bool) {
	principal, ok := ctx.Value(principalContextKey{}).(Principal)
	return principal, ok
}
