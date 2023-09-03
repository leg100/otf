-- name: InsertGithubApp :exec
INSERT INTO github_apps (
    github_app_id,
    webhook_secret,
    pem,
    organization_name
) VALUES (
    pggen.arg('github_app_id'),
    pggen.arg('webhook_secret'),
    pggen.arg('pem'),
    pggen.arg('organization_name')
);

-- name: FindGithubAppByID :one
SELECT *
FROM github_apps
WHERE github_app_id = pggen.arg('github_app_id');

-- name: DeleteGithubAppByID :one
DELETE
FROM github_apps
WHERE github_app_id = pggen.arg('github_app_id')
RETURNING *;
