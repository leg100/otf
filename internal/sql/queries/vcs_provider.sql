-- name: InsertVCSProvider :exec
INSERT INTO vcs_providers (
    vcs_provider_id,
    created_at,
    name,
    cloud,
    token,
    github_app_install_id,
    organization_name
) VALUES (
    pggen.arg('vcs_provider_id'),
    pggen.arg('created_at'),
    pggen.arg('name'),
    pggen.arg('cloud'),
    pggen.arg('token'),
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

-- name: FindVCSProvider :one
SELECT *
FROM vcs_providers
WHERE vcs_provider_id = pggen.arg('vcs_provider_id')
;

-- name: DeleteVCSProviderByID :one
DELETE
FROM vcs_providers
WHERE vcs_provider_id = pggen.arg('vcs_provider_id')
RETURNING vcs_provider_id
;
