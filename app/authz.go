package app

import (
	"context"

	"github.com/leg100/otf"
)

func (a *Application) CanAccessSite(ctx context.Context, action otf.Action) (otf.Subject, error) {
	subj, err := otf.SubjectFromContext(ctx)
	if err != nil {
		return nil, err
	}
	if subj.CanAccessSite(action) {
		return subj, nil
	}
	return nil, otf.ErrAccessNotPermitted
}

func (a *Application) CanAccessOrganization(ctx context.Context, action otf.Action, name string) (otf.Subject, error) {
	subj, err := otf.SubjectFromContext(ctx)
	if err != nil {
		return nil, err
	}
	if subj.CanAccessOrganization(action, name) {
		return subj, nil
	}
	return nil, otf.ErrAccessNotPermitted
}

func (a *Application) CanAccessWorkspace(ctx context.Context, action otf.Action, spec otf.WorkspaceSpec) (otf.Subject, error) {
	subj, err := otf.SubjectFromContext(ctx)
	if err != nil {
		return nil, err
	}
	workspaceID, err := a.db.GetWorkspaceID(ctx, spec)
	if err != nil {
		return nil, err
	}
	organizationName, err := a.db.GetOrganizationNameByWorkspaceID(ctx, workspaceID)
	if err != nil {
		return nil, err
	}
	perms, err := a.db.ListWorkspacePermissions(ctx, otf.WorkspaceSpec{ID: otf.String(workspaceID)})
	if err != nil {
		return nil, err
	}
	if subj.CanAccessWorkspace(action, &otf.WorkspacePolicy{
		OrganizationName: organizationName,
		WorkspaceID:      workspaceID,
		Permissions:      perms,
	}) {
		return subj, nil
	}
	return nil, otf.ErrAccessNotPermitted
}

func (a *Application) CanAccessRun(ctx context.Context, action otf.Action, runID string) (otf.Subject, error) {
	subj, err := otf.SubjectFromContext(ctx)
	if err != nil {
		return nil, err
	}
	workspaceID, err := a.db.GetWorkspaceIDByRunID(ctx, runID)
	if err != nil {
		return nil, err
	}
	organizationName, err := a.db.GetOrganizationNameByWorkspaceID(ctx, workspaceID)
	if err != nil {
		return nil, err
	}
	perms, err := a.db.ListWorkspacePermissions(ctx, otf.WorkspaceSpec{ID: otf.String(workspaceID)})
	if err != nil {
		return nil, err
	}
	if subj.CanAccessWorkspace(action, &otf.WorkspacePolicy{
		OrganizationName: organizationName,
		WorkspaceID:      workspaceID,
		Permissions:      perms,
	}) {
		return subj, nil
	}
	return nil, otf.ErrAccessNotPermitted
}

func (a *Application) CanAccessStateVersion(ctx context.Context, action otf.Action, svID string) (otf.Subject, error) {
	subj, err := otf.SubjectFromContext(ctx)
	if err != nil {
		return nil, err
	}
	workspaceID, err := a.db.GetWorkspaceIDByStateVersionID(ctx, svID)
	if err != nil {
		return nil, err
	}
	organizationName, err := a.db.GetOrganizationNameByWorkspaceID(ctx, workspaceID)
	if err != nil {
		return nil, err
	}
	perms, err := a.db.ListWorkspacePermissions(ctx, otf.WorkspaceSpec{ID: otf.String(workspaceID)})
	if err != nil {
		return nil, err
	}
	if subj.CanAccessWorkspace(action, &otf.WorkspacePolicy{
		OrganizationName: organizationName,
		WorkspaceID:      workspaceID,
		Permissions:      perms,
	}) {
		return subj, nil
	}
	return nil, otf.ErrAccessNotPermitted
}
