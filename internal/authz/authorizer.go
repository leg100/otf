package authz

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/rbac"
	"github.com/leg100/otf/internal/resource"
)

// Authorizer is capable of granting or denying access to resources based on the
// subject contained within the context.
type Authorizer interface {
	CanAccess(ctx context.Context, action rbac.Action, id resource.ID) (Subject, error)
}

type Authy struct {
	logr.Logger
}

func (a *Authy) CanAccess(ctx context.Context, action rbac.Action, organizationName string) (Subject, error) {
	subj, err := SubjectFromContext(ctx)
	if err != nil {
		return nil, err
	}
	if SkipAuthz(ctx) {
		return subj, nil
	}
	if subj.CanAccessOrganization(action, organizationName) {
		return subj, nil
	}
	a.Error(nil, "unauthorized action", "organization", organizationName, "action", action.String(), "subject", subj)
	return nil, internal.ErrAccessNotPermitted
}
