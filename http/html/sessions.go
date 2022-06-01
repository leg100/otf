package html

import (
	"context"
	"fmt"

	"github.com/leg100/otf"
)

// unexported key type prevents collisions
type ctxKey int

const (
	sessionCookie = "session"

	userCtxKey ctxKey = iota
	sessionCtxKey
)

func getCtxUser(ctx context.Context) (*otf.User, error) {
	user, ok := ctx.Value(userCtxKey).(*otf.User)
	if !ok {
		return nil, fmt.Errorf("no user in context")
	}
	return user, nil
}

func getCtxSession(ctx context.Context) (*otf.Session, error) {
	session, ok := ctx.Value(sessionCtxKey).(*otf.Session)
	if !ok {
		return nil, fmt.Errorf("no session in context")
	}
	return session, nil
}
