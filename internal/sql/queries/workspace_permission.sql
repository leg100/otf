-- name: UpsertWorkspacePermission :exec
INSERT INTO workspace_permissions (
    workspace_id,
    team_id,
    role
) VALUES (
    sqlc.arg('workspace_id'),
    sqlc.arg('team_id'),
    sqlc.arg('role')
) ON CONFLICT (workspace_id, team_id) DO UPDATE SET role = sqlc.arg('role')
;

-- name: FindWorkspacePermissionsAndGlobalRemoteState :one
SELECT
    w.global_remote_state,
    array_agg(wp.*)::workspace_permissions[] AS workspace_permissions
FROM workspaces w
LEFT JOIN workspace_permissions wp USING (workspace_id)
WHERE w.workspace_id = sqlc.arg('workspace_id')
GROUP BY w.workspace_id
;

-- name: DeleteWorkspacePermissionByID :exec
DELETE
FROM workspace_permissions
WHERE workspace_id = sqlc.arg('workspace_id')
AND team_id = sqlc.arg('team_id')
;
