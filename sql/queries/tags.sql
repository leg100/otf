-- name: InsertTag :exec
INSERT INTO tags (
    tag_id,
    name,
    organization_name
) VALUES (
    pggen.arg('tag_id'),
    pggen.arg('name'),
    pggen.arg('organization_name')
);

-- name: InsertWorkspaceTag :exec
INSERT INTO workspace_tags (
    tag_id,
    workspace_id
) VALUES (
    pggen.arg('tag_id'),
    pggen.arg('workspace_id')
);

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
;

-- name: DeleteTag :exec
DELETE
FROM tags
WHERE tag_id = pggen.arg('tag_id')
;

-- name: DeleteWorkspaceTag :exec
DELETE
FROM workspace_tags
WHERE workspace_id  = pggen.arg('workspace_id')
AND   tag_id        = pggen.arg('tag_id')
;
