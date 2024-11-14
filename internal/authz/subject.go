package authz

import (
	"context"
	"fmt"

	"github.com/leg100/otf/internal/rbac"
)

// unexported key types prevents collisions
type (
	subjectCtxKeyType string
)

const (
	subjectCtxKey subjectCtxKeyType = "subject"
)

// Subject is an entity that carries out actions on resources.
type Subject interface {
	CanAccess(action rbac.Action, req *AccessRequest) bool
	String() string
}

// AddSubjectToContext adds a subject to a context
func AddSubjectToContext(ctx context.Context, subj Subject) context.Context {
	return context.WithValue(ctx, subjectCtxKey, subj)
}

// SubjectFromContext retrieves a subject from a context
func SubjectFromContext(ctx context.Context) (Subject, error) {
	subj, ok := ctx.Value(subjectCtxKey).(Subject)
	if !ok {
		return nil, fmt.Errorf("no subject in context")
	}
	return subj, nil
}
