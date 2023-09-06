-- name: InsertGithubApp :exec
INSERT INTO github_apps (
    github_app_id,
    webhook_secret,
    private_key,
    organization_name
) VALUES (
    pggen.arg('github_app_id'),
    pggen.arg('webhook_secret'),
    pggen.arg('private_key'),
    pggen.arg('organization_name')
);

-- name: FindGithubAppsByOrganization :one
SELECT *
FROM github_apps
WHERE organization_name = pggen.arg('organization_name');

-- name: FindGithubAppByID :one
SELECT *
FROM github_apps
WHERE github_app_id = pggen.arg('github_app_id');

-- name: DeleteGithubAppByID :one
DELETE
FROM github_apps
WHERE github_app_id = pggen.arg('github_app_id')
RETURNING *;

-- name: InsertGithubAppInstall :exec
INSERT INTO github_app_installs (
    github_app_install_id,
    install_id,
    github_app_id
) VALUES (
    pggen.arg('github_app_install_id'),
    pggen.arg('install_id'),
    pggen.arg('github_app_id')
);

-- name: FindGithubAppInstallByID :one
SELECT *
FROM github_app_installs
JOIN github_apps USING (github_app_id)
WHERE github_app_install_id = pggen.arg('github_app_install_id');

-- name: DeleteGithubAppInstallByID :one
DELETE
FROM github_app_installs
WHERE github_app_install_id = pggen.arg('github_app_install_id')
RETURNING *;
