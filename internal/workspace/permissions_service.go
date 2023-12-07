package workspace

import (
	"context"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/rbac"
)

// GetPolicy retrieves a workspace policy.
//
// NOTE: no authz protects this endpoint because it's used in the process of making
// authz decisions.
func (s *Service) GetPolicy(ctx context.Context, workspaceID string) (internal.WorkspacePolicy, error) {
	return s.db.GetWorkspacePolicy(ctx, workspaceID)
}

func (s *Service) SetPermission(ctx context.Context, workspaceID, teamID string, role rbac.Role) error {
	subject, err := s.CanAccess(ctx, rbac.SetWorkspacePermissionAction, workspaceID)
	if err != nil {
		return err
	}

	if err := s.db.SetWorkspacePermission(ctx, workspaceID, teamID, role); err != nil {
		s.Error(err, "setting workspace permission", "subject", subject, "workspace", workspaceID)
		return err
	}

	s.V(0).Info("set workspace permission", "team_id", teamID, "role", role, "subject", subject, "workspace", workspaceID)

	// TODO: publish event

	return nil
}

func (s *Service) UnsetPermission(ctx context.Context, workspaceID, teamID string) error {
	subject, err := s.CanAccess(ctx, rbac.UnsetWorkspacePermissionAction, workspaceID)
	if err != nil {
		s.Error(err, "unsetting workspace permission", "team_id", teamID, "subject", subject, "workspace", workspaceID)
		return err
	}

	s.V(0).Info("unset workspace permission", "team_id", teamID, "subject", subject, "workspace", workspaceID)
	// TODO: publish event
	return s.db.UnsetWorkspacePermission(ctx, workspaceID, teamID)
}
