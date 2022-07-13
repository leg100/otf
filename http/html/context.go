package html

import (
	"context"
	"fmt"

	"github.com/leg100/otf"
)

// unexported key type prevents collisions
type ctxKey int

const (
	sessionCtxKey ctxKey = iota
	organizationCtxKey
)

func addSessionToContext(ctx context.Context, sess *otf.Session) context.Context {
	return context.WithValue(ctx, sessionCtxKey, sess)
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
