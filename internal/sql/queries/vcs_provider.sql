-- name: InsertVCSProvider :exec
INSERT INTO vcs_providers (
    vcs_provider_id,
    created_at,
    name,
    vcs_kind,
    token,
    github_app_id,
    organization_name
) VALUES (
    pggen.arg('vcs_provider_id'),
    pggen.arg('created_at'),
    pggen.arg('name'),
    pggen.arg('vcs_kind'),
    pggen.arg('token'),
    pggen.arg('github_app_id'),
    pggen.arg('organization_name')
);

-- name: FindVCSProvidersByOrganization :many
SELECT
    v.*,
    (ga.*)::"github_apps" AS github_app,
    (gi.*)::"github_app_installs" AS github_app_install
FROM vcs_providers v
LEFT JOIN (github_app_installs gi JOIN github_apps ga USING (github_app_id)) USING (vcs_provider_id)
WHERE v.organization_name = pggen.arg('organization_name')
;

-- name: FindVCSProviders :many
SELECT
    v.*,
    (ga.*)::"github_apps" AS github_app,
    (gi.*)::"github_app_installs" AS github_app_install
FROM vcs_providers v
LEFT JOIN (github_app_installs gi JOIN github_apps ga USING (github_app_id)) USING (vcs_provider_id)
;

-- name: FindVCSProvidersByGithubAppInstallID :many
SELECT
    v.*,
    (ga.*)::"github_apps" AS github_app,
    (gi.*)::"github_app_installs" AS github_app_install
FROM vcs_providers v
JOIN (github_app_installs gi JOIN github_apps ga USING (github_app_id)) USING (vcs_provider_id)
WHERE gi.install_id = pggen.arg('install_id')
;

-- name: FindVCSProvider :one
SELECT
    v.*,
    (ga.*)::"github_apps" AS github_app,
    (gi.*)::"github_app_installs" AS github_app_install
FROM vcs_providers v
LEFT JOIN (github_app_installs gi JOIN github_apps ga USING (github_app_id)) USING (vcs_provider_id)
WHERE v.vcs_provider_id = pggen.arg('vcs_provider_id')
;

-- name: FindVCSProviderForUpdate :one
SELECT
    v.*,
    (ga.*)::"github_apps" AS github_app,
    (gi.*)::"github_app_installs" AS github_app_install
FROM vcs_providers v
LEFT JOIN (github_app_installs gi JOIN github_apps ga USING (github_app_id)) USING (vcs_provider_id)
WHERE v.vcs_provider_id = pggen.arg('vcs_provider_id')
FOR UPDATE OF v
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
