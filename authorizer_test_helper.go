package otf

import (
	"context"

	"github.com/leg100/otf/rbac"
)

type AllowAllAuthorizer struct {
	User Subject
}

func NewAllowAllAuthorizer() *AllowAllAuthorizer {
	return &AllowAllAuthorizer{
		User: &Superuser{},
	}
}

func (a *AllowAllAuthorizer) CanAccessSite(ctx context.Context, action rbac.Action) (Subject, error) {
	return a.User, nil
}

func (a *AllowAllAuthorizer) CanAccessOrganization(ctx context.Context, action rbac.Action, name string) (Subject, error) {
	return a.User, nil
}

func (a *AllowAllAuthorizer) CanAccessWorkspaceByName(ctx context.Context, action rbac.Action, organization, workspace string) (Subject, error) {
	return a.User, nil
}

func (a *AllowAllAuthorizer) CanAccessWorkspaceByID(ctx context.Context, action rbac.Action, workspaceID string) (Subject, error) {
	return a.User, nil
}

func (a *AllowAllAuthorizer) CanAccessRun(ctx context.Context, action rbac.Action, runID string) (Subject, error) {
	return a.User, nil
}

func (a *AllowAllAuthorizer) CanAccessConfigurationVersion(ctx context.Context, action rbac.Action, cvID string) (Subject, error) {
	return a.User, nil
}

func (a *AllowAllAuthorizer) CanAccessStateVersion(ctx context.Context, action rbac.Action, svID string) (Subject, error) {
	return a.User, nil
}
