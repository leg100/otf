package app

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
	"github.com/leg100/otf/rbac"
)

type authorizer struct {
	db otf.DB
	logr.Logger
}

func (a *authorizer) CanAccessSite(ctx context.Context, action rbac.Action) (otf.Subject, error) {
	subj, err := otf.SubjectFromContext(ctx)
	if err != nil {
		return nil, err
	}
	if subj.CanAccessSite(action) {
		return subj, nil
	}
	a.Error(nil, "unauthorized action", "action", action, "subject", subj)
	return nil, otf.ErrAccessNotPermitted
}

func (a *authorizer) CanAccessOrganization(ctx context.Context, action rbac.Action, name string) (otf.Subject, error) {
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

func (a *authorizer) CanAccessWorkspaceByName(ctx context.Context, action rbac.Action, organization, workspace string) (otf.Subject, error) {
	ws, err := a.db.GetWorkspaceByName(ctx, organization, workspace)
	if err != nil {
		return nil, err
	}
	return a.CanAccessWorkspaceByID(ctx, action, ws.ID())
}

func (a *authorizer) CanAccessWorkspaceByID(ctx context.Context, action rbac.Action, workspaceID string) (otf.Subject, error) {
	subj, err := otf.SubjectFromContext(ctx)
	if err != nil {
		return nil, err
	}
	organization, err := a.db.GetOrganizationNameByWorkspaceID(ctx, workspaceID)
	if err != nil {
		return nil, err
	}
	perms, err := a.db.ListWorkspacePermissions(ctx, workspaceID)
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

func (a *authorizer) CanAccessRun(ctx context.Context, action rbac.Action, runID string) (otf.Subject, error) {
	workspaceID, err := a.db.GetWorkspaceIDByRunID(ctx, runID)
	if err != nil {
		return nil, err
	}
	return a.CanAccessWorkspaceByID(ctx, action, workspaceID)
}

func (a *authorizer) CanAccessConfigurationVersion(ctx context.Context, action rbac.Action, cvID string) (otf.Subject, error) {
	workspaceID, err := a.db.GetWorkspaceIDByCVID(ctx, cvID)
	if err != nil {
		return nil, err
	}
	return a.CanAccessWorkspaceByID(ctx, action, workspaceID)
}

func (a *authorizer) CanAccessStateVersion(ctx context.Context, action rbac.Action, svID string) (otf.Subject, error) {
	workspaceID, err := a.db.GetWorkspaceIDByStateVersionID(ctx, svID)
	if err != nil {
		return nil, err
	}
	return a.CanAccessWorkspaceByID(ctx, action, workspaceID)
}
