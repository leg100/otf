-- name: UpsertWorkspacePermission :exec
INSERT INTO workspace_permissions (
    workspace_id,
    team_id,
    role
) VALUES (
    pggen.arg('workspace_id'),
    pggen.arg('team_id'),
    pggen.arg('role')
) ON CONFLICT (workspace_id, team_id) DO UPDATE SET role = pggen.arg('role');

-- name: FindWorkspacePermissionsByWorkspaceID :many
SELECT *
FROM workspace_permissions
WHERE workspace_id = pggen.arg('workspace_id');

-- name: DeleteWorkspacePermissionByID :exec
DELETE
FROM workspace_permissions
WHERE workspace_id = pggen.arg('workspace_id')
AND team_id = pggen.arg('team_id');
