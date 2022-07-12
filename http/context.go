package http

import (
	"context"
	"fmt"

	"github.com/leg100/otf"
)

// unexported key type prevents collisions
type ctxKey int

const (
	userCtxKey ctxKey = iota
)

func userFromContext(ctx context.Context) (*otf.User, error) {
	user, ok := ctx.Value(userCtxKey).(*otf.User)
	if !ok {
		return nil, fmt.Errorf("no user in context")
	}
	return user, nil
}

func addUserToContext(ctx context.Context, user *otf.User) context.Context {
	return context.WithValue(ctx, userCtxKey, user)
}
