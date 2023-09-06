-- name: InsertGithubApp :exec
INSERT INTO github_apps (
    github_app_id,
    webhook_secret,
    private_key
) VALUES (
    pggen.arg('github_app_id'),
    pggen.arg('webhook_secret'),
    pggen.arg('private_key')
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
