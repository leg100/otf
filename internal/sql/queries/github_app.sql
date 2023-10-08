-- name: InsertGithubApp :exec
INSERT INTO github_apps (
    github_app_id,
    webhook_secret,
    private_key,
    slug,
    organization
) VALUES (
    pggen.arg('github_app_id'),
    pggen.arg('webhook_secret'),
    pggen.arg('private_key'),
    pggen.arg('slug'),
    pggen.arg('organization')
);

-- name: FindGithubApp :one
SELECT *
FROM github_apps;

-- name: DeleteGithubApp :one
DELETE
FROM github_apps
WHERE github_app_id = pggen.arg('github_app_id')
RETURNING *;

-- name: InsertGithubAppInstall :exec
INSERT INTO github_app_installs (
    github_app_id,
    install_id,
    username,
    organization,
    vcs_provider_id
) VALUES (
    pggen.arg('github_app_id'),
    pggen.arg('install_id'),
    pggen.arg('username'),
    pggen.arg('organization'),
    pggen.arg('vcs_provider_id')
);
