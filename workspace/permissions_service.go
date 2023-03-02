package workspace

import (
	"context"

	"github.com/leg100/otf"
	"github.com/leg100/otf/rbac"
)

type permissionsService interface {
	GetPolicy(ctx context.Context, workspaceID string) (otf.WorkspacePolicy, error)

	setPermission(ctx context.Context, workspaceID, team string, role rbac.Role) error
	unsetPermission(ctx context.Context, workspaceID, team string) error
}

func (svc *Service) GetPolicy(ctx context.Context, workspaceID string) (otf.WorkspacePolicy, error) {
	return svc.db.GetWorkspacePolicy(ctx, workspaceID)
}

func (svc *Service) setPermission(ctx context.Context, workspaceID, team string, role rbac.Role) error {
	subject, err := svc.CanAccess(ctx, rbac.SetWorkspacePermissionAction, workspaceID)
	if err != nil {
		return err
	}

	if err := svc.db.SetWorkspacePermission(ctx, workspaceID, team, role); err != nil {
		svc.Error(err, "setting workspace permission", "subject", subject, "workspace", workspaceID)
		return err
	}

	svc.V(0).Info("set workspace permission", "team", team, "role", role, "subject", subject, "workspace", workspaceID)

	// TODO: publish event

	return nil
}

func (svc *Service) unsetPermission(ctx context.Context, workspaceID, team string) error {
	subject, err := svc.CanAccess(ctx, rbac.UnsetWorkspacePermissionAction, workspaceID)
	if err != nil {
		svc.Error(err, "unsetting workspace permission", "team", team, "subject", subject, "workspace", workspaceID)
		return err
	}

	svc.V(0).Info("unset workspace permission", "team", team, "subject", subject, "workspace", workspaceID)
	// TODO: publish event
	return svc.db.UnsetWorkspacePermission(ctx, workspaceID, team)
}
