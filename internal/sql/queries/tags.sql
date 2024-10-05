-- name: InsertTag :exec
INSERT INTO tags (
    tag_id,
    name,
    organization_name
) VALUES (
    sqlc.arg('tag_id'),
    sqlc.arg('name'),
    sqlc.arg('organization_name')
) ON CONFLICT (organization_name, name) DO NOTHING
;

-- name: InsertWorkspaceTag :one
INSERT INTO workspace_tags (
    tag_id,
    workspace_id
) SELECT sqlc.arg('tag_id'), sqlc.arg('workspace_id')
  FROM workspaces w
  JOIN tags t ON (t.organization_name = w.organization_name)
  WHERE w.workspace_id = sqlc.arg('workspace_id')
  AND t.tag_id = sqlc.arg('tag_id')
RETURNING tag_id
;

-- name: InsertWorkspaceTagByName :one
INSERT INTO workspace_tags (
    tag_id,
    workspace_id
) SELECT t.tag_id, sqlc.arg('workspace_id')
  FROM workspaces w
  JOIN tags t ON (t.organization_name = w.organization_name)
  WHERE t.name = sqlc.arg('tag_name')
RETURNING tag_id
;

-- name: FindTags :many
SELECT
    t.*,
    (
        SELECT count(*)
        FROM workspace_tags wt
        WHERE wt.tag_id = t.tag_id
    ) AS instance_count
FROM tags t
WHERE t.organization_name = sqlc.arg('organization_name')
LIMIT sqlc.arg('limit')::int
OFFSET sqlc.arg('offset')::int
;

-- name: FindWorkspaceTags :many
SELECT
    t.*,
    (
        SELECT count(*)
        FROM workspace_tags wt
        WHERE wt.tag_id = t.tag_id
    ) AS instance_count
FROM workspace_tags wt
JOIN tags t USING (tag_id)
WHERE wt.workspace_id = sqlc.arg('workspace_id')
LIMIT sqlc.arg('limit')::int
OFFSET sqlc.arg('offset')::int
;

-- name: FindTagByName :one
SELECT
    t.*,
    (
        SELECT count(*)
        FROM workspace_tags wt
        WHERE wt.tag_id = t.tag_id
    ) AS instance_count
FROM tags t
WHERE t.name = sqlc.arg('name')
AND   t.organization_name = sqlc.arg('organization_name')
;

-- name: FindTagByID :one
SELECT
    t.*,
    (
        SELECT count(*)
        FROM workspace_tags wt
        WHERE wt.tag_id = t.tag_id
    ) AS instance_count
FROM tags t
WHERE t.tag_id = sqlc.arg('tag_id')
AND   t.organization_name = sqlc.arg('organization_name')
;

-- name: CountTags :one
SELECT count(*)
FROM tags t
WHERE t.organization_name = sqlc.arg('organization_name')
;

-- name: CountWorkspaceTags :one
SELECT count(*)
FROM workspace_tags wt
WHERE wt.workspace_id = sqlc.arg('workspace_id')
;

-- name: DeleteTag :one
DELETE
FROM tags
WHERE tag_id            = sqlc.arg('tag_id')
AND   organization_name = sqlc.arg('organization_name')
RETURNING tag_id
;

-- name: DeleteWorkspaceTag :one
DELETE
FROM workspace_tags
WHERE workspace_id  = sqlc.arg('workspace_id')
AND   tag_id        = sqlc.arg('tag_id')
RETURNING tag_id
;
