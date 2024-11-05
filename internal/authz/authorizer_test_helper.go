package authz

import (
	"context"

	"github.com/leg100/otf/internal/rbac"
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

func (a *allowAllAuthorizer) CanAccess(context.Context, rbac.Action, resource.ID) (Subject, error) {
	return a.User, nil
}
