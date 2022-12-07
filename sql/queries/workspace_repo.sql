-- name: InsertWorkspaceRepo :exec
INSERT INTO workspace_repos (
    branch,
    webhook_id,
    vcs_provider_id,
    workspace_id
) VALUES (
    pggen.arg('branch'),
    pggen.arg('webhook_id'),
    pggen.arg('vcs_provider_id'),
    pggen.arg('workspace_id')
);

-- name: UpdateWorkspaceRepoByID :one
UPDATE workspace_repos
SET
    branch = pggen.arg('branch')
WHERE workspace_id = pggen.arg('workspace_id')
RETURNING workspace_id;

-- name: DeleteWorkspaceRepo :one
DELETE
FROM workspace_repos
WHERE workspace_id = pggen.arg('workspace_id')
RETURNING *
;
