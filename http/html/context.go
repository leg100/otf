package html

import (
	"context"
	"fmt"

	"github.com/leg100/otf"
)

// unexported key type prevents collisions
type ctxKey int

const (
	userCtxKey ctxKey = iota
	sessionCtxKey
	organizationCtxKey
)

func userFromContext(ctx context.Context) (*otf.User, error) {
	user, ok := ctx.Value(userCtxKey).(*otf.User)
	if !ok {
		return nil, fmt.Errorf("no user in context")
	}
	return user, nil
}

func sessionFromContext(ctx context.Context) (*otf.Session, error) {
	session, ok := ctx.Value(sessionCtxKey).(*otf.Session)
	if !ok {
		return nil, fmt.Errorf("no session in context")
	}
	return session, nil
}

func newOrganizationContext(ctx context.Context, organizationName string) context.Context {
	return context.WithValue(ctx, organizationCtxKey, organizationName)
}

func organizationFromContext(ctx context.Context) (string, error) {
	name, ok := ctx.Value(organizationCtxKey).(string)
	if !ok {
		return "", fmt.Errorf("no organization in context")
	}
	return name, nil
}
