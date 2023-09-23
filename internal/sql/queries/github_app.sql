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
