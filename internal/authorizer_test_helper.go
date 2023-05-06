package internal

import (
	"context"

	"github.com/leg100/otf/internal/rbac"
)

type allowAllAuthorizer struct {
	User Subject
}

func NewAllowAllAuthorizer() *allowAllAuthorizer {
	return &allowAllAuthorizer{
		User: &Superuser{},
	}
}

func (a *allowAllAuthorizer) CanAccess(context.Context, rbac.Action, string) (Subject, error) {
	return a.User, nil
}
