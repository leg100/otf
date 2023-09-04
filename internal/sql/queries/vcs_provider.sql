-- name: InsertVCSProvider :exec
INSERT INTO vcs_providers (
    vcs_provider_id,
    created_at,
    name,
    cloud,
    token,
    github_app_id,
    organization_name
) VALUES (
    pggen.arg('vcs_provider_id'),
    pggen.arg('created_at'),
    pggen.arg('name'),
    pggen.arg('cloud'),
    pggen.arg('token'),
    pggen.arg('github_app_id'),
    pggen.arg('organization_name')
);

-- name: FindVCSProvidersByOrganization :many
SELECT *
FROM vcs_providers
LEFT JOIN github_apps USING (github_app_id)
WHERE organization_name = pggen.arg('organization_name')
;

-- name: FindVCSProviders :many
SELECT *
FROM vcs_providers
LEFT JOIN github_apps USING (github_app_id)
;

-- name: FindVCSProvider :one
SELECT *
FROM vcs_providers
LEFT JOIN github_apps USING (github_app_id)
WHERE vcs_provider_id = pggen.arg('vcs_provider_id')
;

-- name: DeleteVCSProviderByID :one
DELETE
FROM vcs_providers
WHERE vcs_provider_id = pggen.arg('vcs_provider_id')
RETURNING vcs_provider_id
;

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
