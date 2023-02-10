package auth

import (
	"context"
	"fmt"

	"github.com/leg100/otf"
)

// unexported key type prevents collisions
type ctxKey int

const (
	sessionCtxKey ctxKey = iota
)

// agentFromContext retrieves an agent(-token) from a context
func agentFromContext(ctx context.Context) (*agentToken, error) {
	subj, err := otf.SubjectFromContext(ctx)
	if err != nil {
		return nil, err
	}
	agent, ok := subj.(*agentToken)
	if !ok {
		return nil, fmt.Errorf("subject found in context but it is not an agent")
	}
	return agent, nil
}

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
