package policy

import (
	"context"

	"github.com/leg100/otf"
	"github.com/leg100/otf/rbac"
	"github.com/leg100/otf/sql"
	"github.com/leg100/otf/sql/pggen"
)

type pgdb struct {
	otf.DB
}

func (db *pgdb) setWorkspacePermission(ctx context.Context, workspaceID, team string, role rbac.Role) error {
	_, err := db.UpsertWorkspacePermission(ctx, pggen.UpsertWorkspacePermissionParams{
		WorkspaceID: sql.String(workspaceID),
		TeamName:    sql.String(team),
		Role:        sql.String(role.String()),
	})
	if err != nil {
		return sql.Error(err)
	}
	return nil
}

func (db *pgdb) getWorkspacePolicy(ctx context.Context, workspaceID string) (otf.WorkspacePolicy, error) {
	ws, err := db.FindWorkspaceByID(ctx, sql.String(workspaceID))
	if err != nil {
		return otf.WorkspacePolicy{}, sql.Error(err)
	}
	rows, err := db.FindWorkspacePermissions(ctx, sql.String(workspaceID))
	if err != nil {
		return otf.WorkspacePolicy{}, sql.Error(err)
	}
	policy := otf.WorkspacePolicy{
		Organization: ws.OrganizationName.String,
		WorkspaceID:  ws.WorkspaceID.String,
	}
	for _, perm := range rows {
		role, err := rbac.WorkspaceRoleFromString(perm.Role.String)
		if err != nil {
			return otf.WorkspacePolicy{}, err
		}
		policy.Permissions = append(policy.Permissions, otf.WorkspacePermission{
			Team: perm.TeamName.String,
			Role: role,
		})
	}
	return policy, nil
}

func (db *pgdb) unsetWorkspacePermission(ctx context.Context, workspaceID, team string) error {
	_, err := db.DeleteWorkspacePermissionByID(ctx, sql.String(workspaceID), sql.String(team))
	if err != nil {
		return sql.Error(err)
	}
	return nil
}
