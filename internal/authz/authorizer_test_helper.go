package authz

import (
	"context"

	"github.com/leg100/otf/internal/resource"
)

type allowAllAuthorizer struct {
	User Subject
}

func NewAllowAllAuthorizer() *allowAllAuthorizer {
	return &allowAllAuthorizer{
		User: &Superuser{},
	}
}

func (a *allowAllAuthorizer) Authorize(context.Context, Action, resource.ID, ...CanAccessOption) (Subject, error) {
	return a.User, nil
}

func (a *allowAllAuthorizer) CanAccess(context.Context, Action, resource.ID) bool {
	return true
}
