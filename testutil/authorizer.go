package testutil

import (
	"context"

	"github.com/leg100/otf"
	"github.com/leg100/otf/rbac"
)

type AllowAllAuthorizer struct {
	User otf.Subject
}

func NewAllowAllAuthorizer() *AllowAllAuthorizer {
	return &AllowAllAuthorizer{
		User: &otf.Superuser{},
	}
}

func (a *AllowAllAuthorizer) CanAccessSite(ctx context.Context, action rbac.Action) (otf.Subject, error) {
	return a.User, nil
}

func (a *AllowAllAuthorizer) CanAccessOrganization(ctx context.Context, action rbac.Action, name string) (otf.Subject, error) {
	return a.User, nil
}

func (a *AllowAllAuthorizer) CanAccessWorkspaceByName(ctx context.Context, action rbac.Action, organization, workspace string) (otf.Subject, error) {
	return a.User, nil
}

func (a *AllowAllAuthorizer) CanAccessWorkspaceByID(ctx context.Context, action rbac.Action, workspaceID string) (otf.Subject, error) {
	return a.User, nil
}

func (a *AllowAllAuthorizer) CanAccessRun(ctx context.Context, action rbac.Action, runID string) (otf.Subject, error) {
	return a.User, nil
}

func (a *AllowAllAuthorizer) CanAccessConfigurationVersion(ctx context.Context, action rbac.Action, cvID string) (otf.Subject, error) {
	return a.User, nil
}

func (a *AllowAllAuthorizer) CanAccessStateVersion(ctx context.Context, action rbac.Action, svID string) (otf.Subject, error) {
	return a.User, nil
}
