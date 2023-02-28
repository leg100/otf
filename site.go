package otf

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/leg100/otf/rbac"
)

// SiteAuthorizer authorizes access to site-wide actions
type SiteAuthorizer struct {
	logr.Logger
}

func (a *SiteAuthorizer) CanAccessSite(ctx context.Context, action rbac.Action) (Subject, error) {
	subj, err := SubjectFromContext(ctx)
	if err != nil {
		return nil, err
	}
	if subj.CanAccessSite(action) {
		return subj, nil
	}
	a.Error(nil, "unauthorized action", "action", action, "subject", subj)
	return nil, ErrAccessNotPermitted
}
