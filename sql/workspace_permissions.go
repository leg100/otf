package sql

import (
	"context"
	"fmt"

	"github.com/leg100/otf"
	"github.com/leg100/otf/sql/pggen"
)

func (db *DB) SetWorkspacePermission(ctx context.Context, spec otf.WorkspaceSpec, team string, role otf.Role) error {
	workspaceID, err := db.GetWorkspaceID(ctx, spec)
	if err != nil {
		return databaseError(err)
	}
	_, err = db.UpsertWorkspacePermission(ctx, pggen.UpsertWorkspacePermissionParams{
		WorkspaceID: String(workspaceID),
		TeamName:    String(team),
		Role:        String(role.String()),
	})
	if err != nil {
		return databaseError(err)
	}
	return nil
}

func (db *DB) ListWorkspacePermissions(ctx context.Context, spec otf.WorkspaceSpec) ([]*otf.WorkspacePermission, error) {
	var perms []*otf.WorkspacePermission
	if spec.ID != nil {
		result, err := db.FindWorkspacePermissionsByID(ctx, String(*spec.ID))
		if err != nil {
			return nil, databaseError(err)
		}
		for _, row := range result {
			perm, err := otf.UnmarshalWorkspacePermissionResult(otf.WorkspacePermissionResult(row))
			if err != nil {
				return nil, databaseError(err)
			}
			perms = append(perms, perm)
		}
	} else if spec.Name != nil && spec.OrganizationName != nil {
		result, err := db.FindWorkspacePermissionsByName(ctx, String(*spec.Name), String(*spec.OrganizationName))
		if err != nil {
			return nil, databaseError(err)
		}
		for _, row := range result {
			perm, err := otf.UnmarshalWorkspacePermissionResult(otf.WorkspacePermissionResult(row))
			if err != nil {
				return nil, databaseError(err)
			}
			perms = append(perms, perm)
		}
	} else {
		return nil, fmt.Errorf("invalid workspace spec")
	}
	return perms, nil
}

func (db *DB) UnsetWorkspacePermission(ctx context.Context, spec otf.WorkspaceSpec, team string) error {
	workspaceID, err := db.GetWorkspaceID(ctx, spec)
	if err != nil {
		return databaseError(err)
	}
	_, err = db.DeleteWorkspacePermissionByID(ctx, String(workspaceID), String(team))
	if err != nil {
		return databaseError(err)
	}
	return nil
}
