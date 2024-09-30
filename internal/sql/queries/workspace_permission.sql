-- name: UpsertWorkspacePermission :exec
INSERT INTO workspace_permissions (
    workspace_id,
    team_id,
    role
) VALUES (
    sqlc.arg('workspace_id'),
    sqlc.arg('team_id'),
    sqlc.arg('role')
) ON CONFLICT (workspace_id, team_id) DO UPDATE SET role = sqlc.arg('role');

-- name: FindWorkspacePermissionsByWorkspaceID :many
SELECT *
FROM workspace_permissions
WHERE workspace_id = sqlc.arg('workspace_id');

-- name: DeleteWorkspacePermissionByID :exec
DELETE
FROM workspace_permissions
WHERE workspace_id = sqlc.arg('workspace_id')
AND team_id = sqlc.arg('team_id');
