package http

import (
	"context"
	"fmt"
)

// unexported key type prevents collisions
type ctxKey int

const (
	organizationCtxKey ctxKey = iota
)

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
