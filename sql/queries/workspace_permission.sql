-- name: UpsertWorkspacePermission :exec
INSERT INTO workspace_permissions (
    workspace_id,
    team_id,
    role
) SELECT w.workspace_id, t.team_id, pggen.arg('role')
    FROM teams t
    JOIN organizations o ON t.organization_name = o.name
    JOIN workspaces w ON w.organization_name = o.name
    WHERE t.name = pggen.arg('team_name')
    AND w.workspace_id = pggen.arg('workspace_id')
ON CONFLICT (workspace_id, team_id) DO UPDATE SET role = pggen.arg('role')
;

-- name: FindWorkspacePolicyByID :one
SELECT
    w.organization_name,
    w.workspace_id,
    (
        SELECT array_remove(array_agg(t.*), NULL)
        FROM teams t
        JOIN workspace_permissions wp USING (team_id)
        WHERE wp.workspace_id = w.workspace_id
    ) AS teams,
    (
        SELECT array_remove(array_agg(wp.*), NULL)
        FROM workspace_permissions wp
        WHERE wp.workspace_id = w.workspace_id
    ) AS workspace_permissions
FROM workspaces w
WHERE workspace_id = pggen.arg('workspace_id')
;

-- name: DeleteWorkspacePermissionByID :exec
DELETE
FROM workspace_permissions wp
USING workspaces w, teams t
WHERE wp.team_id = t.team_id
AND wp.workspace_id = pggen.arg('workspace_id')
AND w.workspace_id = wp.workspace_id
AND w.organization_name = t.organization_name
AND t.name = pggen.arg('team_name')
;
