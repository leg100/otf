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
func agentFromContext(ctx context.Context) (*otf.AgentToken, error) {
	subj, err := otf.SubjectFromContext(ctx)
	if err != nil {
		return nil, err
	}
	agent, ok := subj.(*otf.AgentToken)
	if !ok {
		return nil, fmt.Errorf("subject found in context but it is not an agent")
	}
	return agent, nil
}

func addSessionCtx(ctx context.Context, session *Session) context.Context {
	return context.WithValue(ctx, sessionCtxKey, session)
}

func getSessionCtx(ctx context.Context) (*Session, error) {
	session, ok := ctx.Value(sessionCtxKey).(*Session)
	if !ok {
		return nil, fmt.Errorf("no session in context")
	}
	return session, nil
}
