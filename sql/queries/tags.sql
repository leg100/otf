-- name: InsertTag :exec
INSERT INTO tags (
    tag_id,
    name,
    organization_name
) SELECT pggen.arg('tag_id'), pggen.arg('name'), w.organization_name
  FROM workspaces w
  WHERE w.workspace_id = pggen.arg('workspace_id')
ON CONFLICT (name, organization_name) DO NOTHING
;

-- name: InsertWorkspaceTag :one
INSERT INTO workspace_tags (
    tag_id,
    workspace_id
) SELECT pggen.arg('tag_id'), pggen.arg('workspace_id')
  FROM workspaces w
  JOIN tags t ON (t.organization_name = w.organization_name)
  WHERE w.workspace_id = pggen.arg('workspace_id')
  AND t.tag_id = pggen.arg('tag_id')
RETURNING tag_id
;

-- name: InsertWorkspaceTagByName :one
INSERT INTO workspace_tags (
    tag_id,
    workspace_id
) SELECT t.tag_id, pggen.arg('workspace_id')
  FROM workspaces w
  JOIN tags t ON (t.organization_name = w.organization_name)
  WHERE t.name = pggen.arg('tag_name')
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
WHERE t.organization_name = pggen.arg('organization_name')
LIMIT pggen.arg('limit')
OFFSET pggen.arg('offset')
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
WHERE wt.workspace_id = pggen.arg('workspace_id')
LIMIT pggen.arg('limit')
OFFSET pggen.arg('offset')
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
WHERE t.name = pggen.arg('name')
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
WHERE t.tag_id = pggen.arg('tag_id')
;

-- name: CountTags :one
SELECT count(*)
FROM tags t
WHERE t.organization_name = pggen.arg('organization_name')
;

-- name: CountWorkspaceTags :one
SELECT count(*)
FROM workspace_tags wt
WHERE wt.workspace_id = pggen.arg('workspace_id')
;

-- name: DeleteTag :exec
DELETE
FROM tags
WHERE tag_id            = pggen.arg('tag_id')
AND   organization_name = pggen.arg('organization_name')
;

-- name: DeleteWorkspaceTag :exec
DELETE
FROM workspace_tags
WHERE workspace_id  = pggen.arg('workspace_id')
AND   tag_id        = pggen.arg('tag_id')
;
