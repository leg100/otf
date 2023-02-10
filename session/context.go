package session

import (
	"context"
	"fmt"
)

// unexported key type prevents collisions
type ctxKey int

const (
	sessionCtxKey ctxKey = iota
)

func addToContext(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, sessionCtxKey, token)
}

func fromContext(ctx context.Context) (string, error) {
	token, ok := ctx.Value(sessionCtxKey).(string)
	if !ok {
		return "", fmt.Errorf("no session in context")
	}
	return token, nil
}
