package app

import (
	"context"

	"github.com/leg100/otf"
)

type fakeAuthorizer struct {
	subject otf.Subject
}

func (a *fakeAuthorizer) CanAccessSite(ctx context.Context, action otf.Action) (otf.Subject, error) {
	return a.subject, nil
}

func (a *fakeAuthorizer) CanAccessOrganization(ctx context.Context, action otf.Action, name string) (otf.Subject, error) {
	return a.subject, nil
}

func (a *fakeAuthorizer) CanAccessWorkspace(ctx context.Context, action otf.Action, spec otf.WorkspaceSpec) (otf.Subject, error) {
	return a.subject, nil
}

func (a *fakeAuthorizer) CanAccessWorkspaceByID(ctx context.Context, action otf.Action, id string) (otf.Subject, error) {
	return a.subject, nil
}

func (a *fakeAuthorizer) CanAccessRun(ctx context.Context, action otf.Action, runID string) (otf.Subject, error) {
	return a.subject, nil
}

func (a *fakeAuthorizer) CanAccessStateVersion(ctx context.Context, action otf.Action, svID string) (otf.Subject, error) {
	return a.subject, nil
}

func (a *fakeAuthorizer) CanAccessConfigurationVersion(ctx context.Context, action otf.Action, cvID string) (otf.Subject, error) {
	return a.subject, nil
}
