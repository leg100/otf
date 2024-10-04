-- name: InsertModule :exec
INSERT INTO modules (
    module_id,
    created_at,
    updated_at,
    name,
    provider,
    status,
    organization_name
) VALUES (
    sqlc.arg('id'),
    sqlc.arg('created_at'),
    sqlc.arg('updated_at'),
    sqlc.arg('name'),
    sqlc.arg('provider'),
    sqlc.arg('status'),
    sqlc.arg('organization_name')
);

-- name: InsertModuleVersion :one
INSERT INTO module_versions (
    module_version_id,
    version,
    created_at,
    updated_at,
    module_id,
    status
) VALUES (
    sqlc.arg('module_version_id'),
    sqlc.arg('version'),
    sqlc.arg('created_at'),
    sqlc.arg('updated_at'),
    sqlc.arg('module_id'),
    sqlc.arg('status')
)
RETURNING *;

-- name: ListModulesByOrganization :many
SELECT
    m.module_id,
    m.created_at,
    m.updated_at,
    m.name,
    m.provider,
    m.status,
    m.organization_name,
    r.vcs_provider_id,
    r.repo_path,
    (
        SELECT array_agg(v.*)::module_versions[]
        FROM module_versions v
        WHERE v.module_id = m.module_id
        GROUP BY v.module_id
    ) AS module_versions
FROM modules m
LEFT JOIN repo_connections r USING (module_id)
WHERE m.organization_name = sqlc.arg('organization_name')
;

-- name: FindModuleByName :one
SELECT
    m.module_id,
    m.created_at,
    m.updated_at,
    m.name,
    m.provider,
    m.status,
    m.organization_name,
    r.vcs_provider_id,
    r.repo_path,
    (
        SELECT array_agg(v.*)::module_versions[]
        FROM module_versions v
        WHERE v.module_id = m.module_id
        GROUP BY v.module_id
    ) AS module_versions
FROM modules m
LEFT JOIN repo_connections r USING (module_id)
WHERE m.organization_name = sqlc.arg('organization_name')
AND   m.name = sqlc.arg('name')
AND   m.provider = sqlc.arg('provider')
;

-- name: FindModuleByID :one
SELECT
    m.module_id,
    m.created_at,
    m.updated_at,
    m.name,
    m.provider,
    m.status,
    m.organization_name,
    r.vcs_provider_id,
    r.repo_path,
    (
        SELECT array_agg(v.*)::module_versions[]
        FROM module_versions v
        WHERE v.module_id = m.module_id
        GROUP BY v.module_id
    ) AS module_versions
FROM modules m
LEFT JOIN repo_connections r USING (module_id)
WHERE m.module_id = sqlc.arg('id')
;

-- name: FindModuleByConnection :one
SELECT
    m.module_id,
    m.created_at,
    m.updated_at,
    m.name,
    m.provider,
    m.status,
    m.organization_name,
    r.vcs_provider_id,
    r.repo_path,
    (
        SELECT array_agg(v.*)::module_versions[]
        FROM module_versions v
        WHERE v.module_id = m.module_id
        GROUP BY v.module_id
    ) AS module_versions
FROM modules m
JOIN repo_connections r USING (module_id)
WHERE r.vcs_provider_id = sqlc.arg('vcs_provider_id')
AND   r.repo_path = sqlc.arg('repo_path')
;

-- name: FindModuleByModuleVersionID :one
SELECT
    m.module_id,
    m.created_at,
    m.updated_at,
    m.name,
    m.provider,
    m.status,
    m.organization_name,
    r.vcs_provider_id,
    r.repo_path,
    (
        SELECT array_agg(v.*)::module_versions[]
        FROM module_versions v
        WHERE v.module_id = m.module_id
        GROUP BY v.module_id
    ) AS module_versions
FROM modules m
JOIN module_versions mv USING (module_id)
LEFT JOIN repo_connections r USING (module_id)
WHERE mv.module_version_id = sqlc.arg('module_version_id')
;

-- name: UpdateModuleStatusByID :one
UPDATE modules
SET status = sqlc.arg('status')
WHERE module_id = sqlc.arg('module_id')
RETURNING module_id
;

-- name: InsertModuleTarball :one
INSERT INTO module_tarballs (
    tarball,
    module_version_id
) VALUES (
    sqlc.arg('tarball'),
    sqlc.arg('module_version_id')
)
RETURNING module_version_id;

-- name: FindModuleTarball :one
SELECT tarball
FROM module_tarballs
WHERE module_version_id = sqlc.arg('module_version_id')
;

-- name: UpdateModuleVersionStatusByID :one
UPDATE module_versions
SET
    status = sqlc.arg('status'),
    status_error = sqlc.arg('status_error')
WHERE module_version_id = sqlc.arg('module_version_id')
RETURNING *
;

-- name: DeleteModuleByID :one
DELETE
FROM modules
WHERE module_id = sqlc.arg('module_id')
RETURNING module_id
;

-- name: DeleteModuleVersionByID :one
DELETE
FROM module_versions
WHERE module_version_id = sqlc.arg('module_version_id')
RETURNING module_version_id
;
