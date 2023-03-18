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

// GetPolicy retrieves a workspace policy.
//
// NOTE: no authz protects this endpoint because it's used in the process of making
// authz decisions.
func (s *service) GetPolicy(ctx context.Context, workspaceID string) (otf.WorkspacePolicy, error) {
	return s.db.GetWorkspacePolicy(ctx, workspaceID)
}

func (s *service) setPermission(ctx context.Context, workspaceID, team string, role rbac.Role) error {
	subject, err := s.CanAccess(ctx, rbac.SetWorkspacePermissionAction, workspaceID)
	if err != nil {
		return err
	}

	if err := s.db.SetWorkspacePermission(ctx, workspaceID, team, role); err != nil {
		s.Error(err, "setting workspace permission", "subject", subject, "workspace", workspaceID)
		return err
	}

	s.V(0).Info("set workspace permission", "team", team, "role", role, "subject", subject, "workspace", workspaceID)

	// TODO: publish event

	return nil
}

func (s *service) unsetPermission(ctx context.Context, workspaceID, team string) error {
	subject, err := s.CanAccess(ctx, rbac.UnsetWorkspacePermissionAction, workspaceID)
	if err != nil {
		s.Error(err, "unsetting workspace permission", "team", team, "subject", subject, "workspace", workspaceID)
		return err
	}

	s.V(0).Info("unset workspace permission", "team", team, "subject", subject, "workspace", workspaceID)
	// TODO: publish event
	return s.db.UnsetWorkspacePermission(ctx, workspaceID, team)
}
