-- name: InsertRepoConnection :exec
INSERT INTO repo_connections (
    vcs_provider_id,
    repo_path,
    workspace_id,
    module_id
) VALUES (
    sqlc.arg('vcs_provider_id'),
    sqlc.arg('repo_path'),
    sqlc.arg('workspace_id'),
    sqlc.arg('module_id')
);

-- name: DeleteWorkspaceConnectionByID :one
DELETE
FROM repo_connections
WHERE workspace_id = sqlc.arg('workspace_id')
RETURNING *;

-- name: DeleteModuleConnectionByID :one
DELETE
FROM repo_connections
WHERE module_id = sqlc.arg('module_id')
RETURNING *;
