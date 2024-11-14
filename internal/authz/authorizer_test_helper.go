package authz

import (
	"context"
)

type allowAllAuthorizer struct {
	User Subject
}

func NewAllowAllAuthorizer() *allowAllAuthorizer {
	return &allowAllAuthorizer{
		User: &Superuser{},
	}
}

func (a *allowAllAuthorizer) Authorize(context.Context, Action, *AccessRequest, ...CanAccessOption) (Subject, error) {
	return a.User, nil
}

func (a *allowAllAuthorizer) CanAccess(context.Context, Action, *AccessRequest) bool {
	return true
}
