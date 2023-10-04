-- name: InsertVCSProvider :exec
INSERT INTO vcs_providers (
    vcs_provider_id,
    created_at,
    name,
    vcs_kind,
    token,
    github_app_id,
    github_app_install_id,
    organization_name
) VALUES (
    pggen.arg('vcs_provider_id'),
    pggen.arg('created_at'),
    pggen.arg('name'),
    pggen.arg('vcs_kind'),
    pggen.arg('token'),
    pggen.arg('github_app_id'),
    pggen.arg('github_app_install_id'),
    pggen.arg('organization_name')
);

-- name: FindVCSProvidersByOrganization :many
SELECT *
FROM vcs_providers
WHERE organization_name = pggen.arg('organization_name')
;

-- name: FindVCSProviders :many
SELECT *
FROM vcs_providers
;

-- name: FindVCSProvidersByGithubAppInstallID :many
SELECT *
FROM vcs_providers
WHERE github_app_install_id = pggen.arg('github_app_install_id')
;

-- name: FindVCSProvider :one
SELECT *
FROM vcs_providers
WHERE vcs_provider_id = pggen.arg('vcs_provider_id')
;

-- name: FindVCSProviderForUpdate :one
SELECT *
FROM vcs_providers
WHERE vcs_provider_id = pggen.arg('vcs_provider_id')
FOR UPDATE
;

-- name: UpdateVCSProvider :one
UPDATE vcs_providers
SET name = pggen.arg('name'), token = pggen.arg('token')
WHERE vcs_provider_id = pggen.arg('vcs_provider_id')
RETURNING *
;

-- name: DeleteVCSProviderByID :one
DELETE
FROM vcs_providers
WHERE vcs_provider_id = pggen.arg('vcs_provider_id')
RETURNING vcs_provider_id
;
