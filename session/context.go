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

func addToContext(ctx context.Context, sess *Session) context.Context {
	return context.WithValue(ctx, sessionCtxKey, sess)
}

func fromContext(ctx context.Context) (*Session, error) {
	session, ok := ctx.Value(sessionCtxKey).(*Session)
	if !ok {
		return nil, fmt.Errorf("no session in context")
	}
	return session, nil
}
