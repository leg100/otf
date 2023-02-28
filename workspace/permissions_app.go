package workspace

import (
	"context"

	"github.com/leg100/otf"
	"github.com/leg100/otf/rbac"
)

type permissionsApp interface {
	GetPolicy(ctx context.Context, workspaceID string) (otf.WorkspacePolicy, error)

	setPermission(ctx context.Context, workspaceID, team string, role rbac.Role) error
	unsetPermission(ctx context.Context, workspaceID, team string) error
}

func (a *app) GetPolicy(ctx context.Context, workspaceID string) (otf.WorkspacePolicy, error) {
	return a.db.GetWorkspacePolicy(ctx, workspaceID)
}

func (a *app) setPermission(ctx context.Context, workspaceID, team string, role rbac.Role) error {
	subject, err := a.CanAccessWorkspaceByID(ctx, rbac.SetWorkspacePermissionAction, workspaceID)
	if err != nil {
		return err
	}

	if err := a.db.SetWorkspacePermission(ctx, workspaceID, team, role); err != nil {
		a.Error(err, "setting workspace permission", "subject", subject, "workspace", workspaceID)
		return err
	}

	a.V(0).Info("set workspace permission", "team", team, "role", role, "subject", subject, "workspace", workspaceID)

	// TODO: publish event

	return nil
}

func (a *app) unsetPermission(ctx context.Context, workspaceID, team string) error {
	subject, err := a.CanAccessWorkspaceByID(ctx, rbac.UnsetWorkspacePermissionAction, workspaceID)
	if err != nil {
		a.Error(err, "unsetting workspace permission", "team", team, "subject", subject, "workspace", workspaceID)
		return err
	}

	a.V(0).Info("unset workspace permission", "team", team, "subject", subject, "workspace", workspaceID)
	// TODO: publish event
	return a.db.UnsetWorkspacePermission(ctx, workspaceID, team)
}
