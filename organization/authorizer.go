package organization

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
	"github.com/leg100/otf/rbac"
)

// Authorizer authorizes access to an organization
type Authorizer struct {
	logr.Logger
}

func (a *Authorizer) CanAccessOrganization(ctx context.Context, action rbac.Action, name string) (otf.Subject, error) {
	subj, err := otf.SubjectFromContext(ctx)
	if err != nil {
		return nil, err
	}
	if subj.CanAccessOrganization(action, name) {
		return subj, nil
	}
	a.Error(nil, "unauthorized action", "organization", name, "action", action, "subject", subj)
	return nil, otf.ErrAccessNotPermitted
}
