-- name: InsertTag :exec
INSERT INTO tags (
    tag_id,
    name,
    organization_name
) VALUES (
    pggen.arg('tag_id'),
    pggen.arg('name'),
    pggen.arg('organization_name')
) ON CONFLICT (name, organization_name) DO NOTHING;

-- name: InsertWorkspaceTag :exec
INSERT INTO workspace_tags (
    tag_id,
    workspace_id
) VALUES (
    pggen.arg('tag_id'),
    pggen.arg('workspace_id')
);

-- name: InsertWorkspaceTagByName :exec
INSERT INTO workspace_tags (
    tag_id,
    workspace_id
) SELECT tag_id, pggen.arg('workspace_id')
    FROM tags
    WHERE name = pggen.arg('tag_name')
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

-- name: CountTags :one
SELECT count(*)
FROM tags t
WHERE t.organization_name = pggen.arg('organization_name')
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
