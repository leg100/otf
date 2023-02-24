package workspace

import (
	"context"

	"github.com/jackc/pgtype"
	"github.com/leg100/otf"
	"github.com/leg100/otf/rbac"
	"github.com/leg100/otf/sql"
	"github.com/leg100/otf/sql/pggen"
)

func (db *pgdb) SetWorkspacePermission(ctx context.Context, workspaceID, team string, role rbac.Role) error {
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

func (db *pgdb) ListWorkspacePermissions(ctx context.Context, workspaceID string) (otf.WorkspacePolicy, error) {
	result, err := db.FindWorkspacePermissionsByID(ctx, sql.String(workspaceID))
	if err != nil {
		return otf.WorkspacePolicy{}, sql.Error(err)
	}
	policy := otf.WorkspacePolicy{
		Organization: result.OrganizationName.String,
		WorkspaceID: result.WorkspaceID.String,
	}
	for _, row := range result {
		perm, err := permissionRow(row).toPermission()
		if err != nil {
			return otf.WorkspacePolicy{}, sql.Error(err)
		}
		perms = append(perms, perm)
	}
	return perms, nil
}

func (db *pgdb) UnsetWorkspacePermission(ctx context.Context, workspaceID, team string) error {
	_, err := db.DeleteWorkspacePermissionByID(ctx, sql.String(workspaceID), sql.String(team))
	if err != nil {
		return sql.Error(err)
	}
	return nil
}

// permissionRow represents the result of a database query for a
// workspace permission.
type permissionRow struct {
	Role         pgtype.Text          `json:"role"`
	Team         *pggen.Teams         `json:"team"`
	Organization *pggen.Organizations `json:"organization"`
}

func (row permissionRow) toPermission() (otf.WorkspacePermission, error) {
	role, err := rbac.WorkspaceRoleFromString(row.Role.String)
	if err != nil {
		return otf.WorkspacePermission{}, err
	}
	return otf.WorkspacePermission{
		Role:   role,
		TeamID: row.Team.TeamID.String,
	}, nil
}
