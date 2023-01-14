package app

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
)

// Authorizer is capable of granting or denying access to resources based on the
// subject contained within the context.
type Authorizer interface {
	CanAccessSite(ctx context.Context, action otf.Action) (otf.Subject, error)
	CanAccessOrganization(ctx context.Context, action otf.Action, name string) (otf.Subject, error)
	CanAccessWorkspace(ctx context.Context, action otf.Action, spec otf.WorkspaceSpec) (otf.Subject, error)
	CanAccessWorkspaceByID(ctx context.Context, action otf.Action, workspaceID string) (otf.Subject, error)
	CanAccessRun(ctx context.Context, action otf.Action, runID string) (otf.Subject, error)
	CanAccessStateVersion(ctx context.Context, action otf.Action, svID string) (otf.Subject, error)
	CanAccessConfigurationVersion(ctx context.Context, action otf.Action, cvID string) (otf.Subject, error)
}

type authorizer struct {
	db otf.DB
	logr.Logger
}

func (a *authorizer) CanAccessSite(ctx context.Context, action otf.Action) (otf.Subject, error) {
	subj, err := otf.SubjectFromContext(ctx)
	if err != nil {
		return nil, err
	}
	if subj.CanAccessSite(action) {
		return subj, nil
	}
	return nil, otf.ErrAccessNotPermitted
}

func (a *authorizer) CanAccessOrganization(ctx context.Context, action otf.Action, name string) (otf.Subject, error) {
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

func (a *authorizer) CanAccessWorkspace(ctx context.Context, action otf.Action, spec otf.WorkspaceSpec) (otf.Subject, error) {
	workspaceID, err := a.db.GetWorkspaceID(ctx, spec)
	if err != nil {
		return nil, err
	}
	return a.CanAccessWorkspaceByID(ctx, action, workspaceID)
}

func (a *authorizer) CanAccessRun(ctx context.Context, action otf.Action, runID string) (otf.Subject, error) {
	workspaceID, err := a.db.GetWorkspaceIDByRunID(ctx, runID)
	if err != nil {
		return nil, err
	}
	return a.CanAccessWorkspaceByID(ctx, action, workspaceID)
}

func (a *authorizer) CanAccessConfigurationVersion(ctx context.Context, action otf.Action, cvID string) (otf.Subject, error) {
	workspaceID, err := a.db.GetWorkspaceIDByCVID(ctx, cvID)
	if err != nil {
		return nil, err
	}
	return a.CanAccessWorkspaceByID(ctx, action, workspaceID)
}

func (a *authorizer) CanAccessStateVersion(ctx context.Context, action otf.Action, svID string) (otf.Subject, error) {
	workspaceID, err := a.db.GetWorkspaceIDByStateVersionID(ctx, svID)
	if err != nil {
		return nil, err
	}
	return a.CanAccessWorkspaceByID(ctx, action, workspaceID)
}

func (a *authorizer) CanAccessWorkspaceByID(ctx context.Context, action otf.Action, workspaceID string) (otf.Subject, error) {
	subj, err := otf.SubjectFromContext(ctx)
	if err != nil {
		return nil, err
	}
	organization, err := a.db.GetOrganizationNameByWorkspaceID(ctx, workspaceID)
	if err != nil {
		return nil, err
	}
	perms, err := a.db.ListWorkspacePermissions(ctx, otf.WorkspaceSpec{ID: otf.String(workspaceID)})
	if err != nil {
		return nil, err
	}
	if subj.CanAccessWorkspace(action, &otf.WorkspacePolicy{
		Organization: organization,
		WorkspaceID:  workspaceID,
		Permissions:  perms,
	}) {
		return subj, nil
	}
	a.Error(nil, "unauthorized action", "workspace", workspaceID, "organization", organization, "action", action, "subject", subj)
	return nil, otf.ErrAccessNotPermitted
}
