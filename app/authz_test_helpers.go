package app

import (
	"context"

	"github.com/leg100/otf"
	"github.com/leg100/otf/rbac"
)

type fakeAuthorizer struct {
	subject otf.Subject
}

func (a *fakeAuthorizer) CanAccessSite(ctx context.Context, action rbac.Action) (otf.Subject, error) {
	return a.subject, nil
}

func (a *fakeAuthorizer) CanAccessOrganization(ctx context.Context, action rbac.Action, name string) (otf.Subject, error) {
	return a.subject, nil
}

func (a *fakeAuthorizer) CanAccessWorkspaceByID(ctx context.Context, action rbac.Action, id string) (otf.Subject, error) {
	return a.subject, nil
}

func (a *fakeAuthorizer) CanAccessWorkspaceByName(ctx context.Context, action rbac.Action, organization, workspace string) (otf.Subject, error) {
	return a.subject, nil
}

func (a *fakeAuthorizer) CanAccessRun(ctx context.Context, action rbac.Action, runID string) (otf.Subject, error) {
	return a.subject, nil
}

func (a *fakeAuthorizer) CanAccessStateVersion(ctx context.Context, action rbac.Action, svID string) (otf.Subject, error) {
	return a.subject, nil
}

func (a *fakeAuthorizer) CanAccessConfigurationVersion(ctx context.Context, action rbac.Action, cvID string) (otf.Subject, error) {
	return a.subject, nil
}
