package otf

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/leg100/otf/rbac"
)

// Authorizer is capable of granting or denying access to resources based on the
// subject contained within the context.
type Authorizer interface {
	CanAccessRun(ctx context.Context, action rbac.Action, runID string) (Subject, error)
	CanAccessStateVersion(ctx context.Context, action rbac.Action, svID string) (Subject, error)
	CanAccessConfigurationVersion(ctx context.Context, action rbac.Action, cvID string) (Subject, error)
}

type SiteAuthorizer interface {
	CanAccessSite(ctx context.Context, action rbac.Action) (Subject, error)
}

type OrganizationAuthorizer interface {
	CanAccessOrganization(ctx context.Context, action rbac.Action, name string) (Subject, error)
}

type WorkspaceAuthorizer interface {
	CanAccessWorkspaceByName(ctx context.Context, action rbac.Action, organization, workspace string) (Subject, error)
	CanAccessWorkspaceByID(ctx context.Context, action rbac.Action, workspaceID string) (Subject, error)
}

type siteAuthorizer struct {
	logr.Logger
}

func NewSiteAuthorizer(logger logr.Logger) *siteAuthorizer {
	return &siteAuthorizer{logger}
}

func (a *siteAuthorizer) CanAccessSite(ctx context.Context, action rbac.Action) (Subject, error) {
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
