package workspace

import (
	"context"

	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/rbac"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/sql/sqlc"
)

func (db *pgdb) SetWorkspacePermission(ctx context.Context, workspaceID, teamID resource.ID, role rbac.Role) error {
	err := db.Querier(ctx).UpsertWorkspacePermission(ctx, sqlc.UpsertWorkspacePermissionParams{
		WorkspaceID: workspaceID,
		TeamID:      teamID,
		Role:        sql.String(role.String()),
	})
	if err != nil {
		return sql.Error(err)
	}
	return nil
}

func (db *pgdb) GetWorkspacePolicy(ctx context.Context, workspaceID resource.ID) (authz.WorkspacePolicy, error) {
	q := db.Querier(ctx)

	// Retrieve not only permissions but the workspace too, so that:
	// (1) we ensure that workspace exists and return not found if not
	// (2) we retrieve the name of the organization, which is part of a policy
	ws, err := q.FindWorkspaceByID(ctx, workspaceID)
	if err != nil {
		return authz.WorkspacePolicy{}, sql.Error(err)
	}
	perms, err := q.FindWorkspacePermissionsByWorkspaceID(ctx, workspaceID)
	if err != nil {
		return authz.WorkspacePolicy{}, sql.Error(err)
	}

	policy := authz.WorkspacePolicy{
		Organization:      ws.OrganizationName.String,
		WorkspaceID:       workspaceID,
		GlobalRemoteState: ws.GlobalRemoteState.Bool,
	}
	for _, perm := range perms {
		role, err := rbac.WorkspaceRoleFromString(perm.Role.String)
		if err != nil {
			return authz.WorkspacePolicy{}, err
		}
		policy.Permissions = append(policy.Permissions, authz.WorkspacePermission{
			TeamID: resource.ID{Kind: resource.TeamKind, ID: perm.TeamID.String},
			Role:   role,
		})
	}
	return policy, nil
}

func (db *pgdb) UnsetWorkspacePermission(ctx context.Context, workspaceID, teamID resource.ID) error {
	err := db.Querier(ctx).DeleteWorkspacePermissionByID(ctx, sqlc.DeleteWorkspacePermissionByIDParams{
		WorkspaceID: workspaceID,
		TeamID:      teamID,
	})
	if err != nil {
		return sql.Error(err)
	}
	return nil
}
