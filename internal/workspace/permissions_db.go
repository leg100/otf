package workspace

import (
	"context"

	"github.com/jackc/pgx/v4"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/rbac"
	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/sql/pggen"
)

func (db *pgdb) SetWorkspacePermission(ctx context.Context, workspaceID, team string, role rbac.Role) error {
	_, err := db.Conn(ctx).UpsertWorkspacePermission(ctx, pggen.UpsertWorkspacePermissionParams{
		WorkspaceID: sql.String(workspaceID),
		TeamName:    sql.String(team),
		Role:        sql.String(role.String()),
	})
	if err != nil {
		return sql.Error(err)
	}
	return nil
}

func (db *pgdb) GetWorkspacePolicy(ctx context.Context, workspaceID string) (internal.WorkspacePolicy, error) {
	q := db.Conn(ctx)
	batch := &pgx.Batch{}

	// Retrieve not only permissions but the workspace too, so that:
	// (1) we ensure that workspace exists and return not found if not
	// (2) we retrieve the name of the organization, which is part of a policy
	q.FindWorkspaceByIDBatch(batch, sql.String(workspaceID))
	q.FindWorkspacePermissionsByWorkspaceIDBatch(batch, sql.String(workspaceID))
	results := db.SendBatch(ctx, batch)
	defer results.Close()

	ws, err := q.FindWorkspaceByIDScan(results)
	if err != nil {
		return internal.WorkspacePolicy{}, sql.Error(err)
	}
	perms, err := q.FindWorkspacePermissionsByWorkspaceIDScan(results)
	if err != nil {
		return internal.WorkspacePolicy{}, sql.Error(err)
	}

	policy := internal.WorkspacePolicy{
		Organization:      ws.OrganizationName.String,
		WorkspaceID:       workspaceID,
		GlobalRemoteState: ws.GlobalRemoteState.Bool,
	}
	for _, perm := range perms {
		role, err := rbac.WorkspaceRoleFromString(perm.Role.String)
		if err != nil {
			return internal.WorkspacePolicy{}, err
		}
		policy.Permissions = append(policy.Permissions, internal.WorkspacePermission{
			Team:   perm.Team.Name.String,
			TeamID: perm.Team.TeamID.String,
			Role:   role,
		})
	}
	return policy, nil
}

func (db *pgdb) UnsetWorkspacePermission(ctx context.Context, workspaceID, team string) error {
	_, err := db.Conn(ctx).DeleteWorkspacePermissionByID(ctx, sql.String(workspaceID), sql.String(team))
	if err != nil {
		return sql.Error(err)
	}
	return nil
}
