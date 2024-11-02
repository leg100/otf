package organization

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/rbac"
)

// Authorizer authorizes access to an organization
type Authorizer struct {
	logr.Logger
}

func (a *Authorizer) CanAccess(ctx context.Context, action rbac.Action, name string) (authz.Subject, error) {
	subj, err := authz.SubjectFromContext(ctx)
	if err != nil {
		return nil, err
	}
	if authz.SkipAuthz(ctx) {
		return subj, nil
	}
	if subj.CanAccessOrganization(action, name) {
		return subj, nil
	}
	a.Error(nil, "unauthorized action", "organization", name, "action", action.String(), "subject", subj)
	return nil, internal.ErrAccessNotPermitted
}
