-- name: UpsertWorkspacePermission :exec
INSERT INTO workspace_permissions (
    workspace_id,
    team_id,
    role
) SELECT w.workspace_id, t.team_id, pggen.arg('role')
    FROM teams t
    JOIN organizations o ON t.organization_id = o.organization_id
    JOIN workspaces w ON w.organization_id = o.organization_id
    WHERE t.name = pggen.arg('team_name')
    AND w.workspace_id = pggen.arg('workspace_id')
ON CONFLICT (workspace_id, team_id) DO UPDATE SET role = pggen.arg('role')
;

-- name: FindWorkspacePermissionsByID :many
SELECT
    perms.role,
    (teams.*)::"teams" AS team
FROM workspace_permissions perms
JOIN teams USING (team_id)
WHERE perms.workspace_id = pggen.arg('workspace_id')
;

-- name: FindWorkspacePermissionsByName :many
SELECT
    perms.role,
    (t.*)::"teams" AS team
FROM workspace_permissions perms
JOIN teams t USING (team_id)
JOIN workspaces w USING (workspace_id)
JOIN organizations o ON o.organization_id = w.organization_id
WHERE w.name = pggen.arg('workspace_name')
AND o.name = pggen.arg('organization_name')
;

-- name: DeleteWorkspacePermissionByID :exec
DELETE
FROM workspace_permissions p
USING workspaces w, teams t
WHERE p.team_id = t.team_id
AND w.workspace_id = pggen.arg('workspace_id')
AND t.organization_id = w.organization_id
AND t.name = pggen.arg('team_name')
;

-- name: DeleteWorkspacePermissionByName :exec
DELETE
FROM workspace_permissions p
USING organizations o, workspaces w, teams t
WHERE p.team_id = t.team_id
AND p.workspace_id = w.workspace_id
AND w.name = pggen.arg('workspace_name')
AND o.name = pggen.arg('organization_name')
AND t.name = pggen.arg('team_name')
;
