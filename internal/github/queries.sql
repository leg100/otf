-- name: InsertGithubApp :exec
INSERT INTO github_apps (
    github_app_id,
    webhook_secret,
    private_key,
    slug,
    organization
) VALUES (
    sqlc.arg('github_app_id'),
    sqlc.arg('webhook_secret'),
    sqlc.arg('private_key'),
    sqlc.arg('slug'),
    sqlc.arg('organization')
);

-- name: FindGithubApp :one
SELECT *
FROM github_apps;

-- name: DeleteGithubApp :one
DELETE
FROM github_apps
WHERE github_app_id = sqlc.arg('github_app_id')
RETURNING *;

-- name: InsertGithubAppInstall :exec
INSERT INTO github_app_installs (
    github_app_id,
    install_id,
    username,
    organization,
    vcs_provider_id
) VALUES (
    sqlc.arg('github_app_id'),
    sqlc.arg('install_id'),
    sqlc.arg('username'),
    sqlc.arg('organization'),
    sqlc.arg('vcs_provider_id')
);
