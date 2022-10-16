package app

import (
	"context"

	"github.com/leg100/otf"
)

func (a *Application) SetWorkspacePermission(ctx context.Context, spec otf.WorkspaceSpec, team string, role otf.WorkspaceRole) error {
	subject, err := a.CanAccessWorkspace(ctx, otf.SetWorkspacePermissionAction, spec)
	if err != nil {
		return err
	}

	if err := a.db.SetWorkspacePermission(ctx, spec, team, role); err != nil {
		a.Error(err, "setting workspace permission", append(spec.LogFields(), "subject", subject)...)
		return err
	}

	a.V(0).Info("set workspace permission", append(spec.LogFields(), "team", team, "role", role, "subject", subject)...)

	// TODO: publish event

	return nil
}

func (a *Application) ListWorkspacePermissions(ctx context.Context, spec otf.WorkspaceSpec) ([]*otf.WorkspacePermission, error) {
	return a.db.ListWorkspacePermissions(ctx, spec)
}

func (a *Application) UnsetWorkspacePermission(ctx context.Context, spec otf.WorkspaceSpec, team string) error {
	// TODO: publish event
	return a.db.UnsetWorkspacePermission(ctx, spec, team)
}
