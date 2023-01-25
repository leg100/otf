package sql

import (
	"context"

	"github.com/leg100/otf"
	"github.com/leg100/otf/rbac"
	"github.com/leg100/otf/sql/pggen"
)

func (db *DB) SetWorkspacePermission(ctx context.Context, workspaceID, team string, role rbac.Role) error {
	_, err := db.UpsertWorkspacePermission(ctx, pggen.UpsertWorkspacePermissionParams{
		WorkspaceID: String(workspaceID),
		TeamName:    String(team),
		Role:        String(role.String()),
	})
	if err != nil {
		return Error(err)
	}
	return nil
}

func (db *DB) ListWorkspacePermissions(ctx context.Context, workspaceID string) ([]*otf.WorkspacePermission, error) {
	result, err := db.FindWorkspacePermissionsByID(ctx, String(workspaceID))
	if err != nil {
		return nil, Error(err)
	}
	var perms []*otf.WorkspacePermission
	for _, row := range result {
		perm, err := otf.UnmarshalWorkspacePermissionResult(otf.WorkspacePermissionResult(row))
		if err != nil {
			return nil, Error(err)
		}
		perms = append(perms, perm)
	}
	return perms, nil
}

func (db *DB) UnsetWorkspacePermission(ctx context.Context, workspaceID, team string) error {
	_, err := db.DeleteWorkspacePermissionByID(ctx, String(workspaceID), String(team))
	if err != nil {
		return Error(err)
	}
	return nil
}
