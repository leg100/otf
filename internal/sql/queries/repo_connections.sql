-- name: InsertRepoConnection :exec
INSERT INTO repo_connections (
    vcs_provider_id,
    repo_path,
    workspace_id,
    module_id
) VALUES (
    pggen.arg('vcs_provider_id'),
    pggen.arg('repo_path'),
    pggen.arg('workspace_id'),
    pggen.arg('module_id')
);

-- name: DeleteWorkspaceConnectionByID :one
DELETE
FROM repo_connections
WHERE workspace_id = pggen.arg('workspace_id')
RETURNING *;

-- name: DeleteModuleConnectionByID :one
DELETE
FROM repo_connections
WHERE module_id = pggen.arg('module_id')
RETURNING *;
