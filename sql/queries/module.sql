-- name: InsertModule :exec
INSERT INTO modules (
    module_id,
    created_at,
    updated_at,
    name,
    provider,
    organization_id
) VALUES (
    pggen.arg('id'),
    pggen.arg('created_at'),
    pggen.arg('updated_at'),
    pggen.arg('name'),
    pggen.arg('provider'),
    pggen.arg('organization_id')
);

-- name: InsertModuleVersion :one
INSERT INTO module_versions (
    version,
    created_at,
    updated_at,
    module_id
) VALUES (
    pggen.arg('version'),
    pggen.arg('created_at'),
    pggen.arg('updated_at'),
    pggen.arg('module_id')
)
RETURNING *;

-- name: ListModulesByOrganization :many
SELECT
    m.module_id,
    m.created_at,
    m.updated_at,
    m.name,
    m.provider,
    (o.*)::"organizations" AS organization
FROM modules m
JOIN organizations o USING (organization_id)
WHERE o.name = pggen.arg('organization_name')
;

-- name: FindModuleByName :one
SELECT
    m.module_id,
    m.created_at,
    m.updated_at,
    m.name,
    m.provider,
    (o.*)::"organizations" AS organization
FROM modules m
JOIN organizations o USING (organization_id)
WHERE o.name = pggen.arg('organizaton_name')
AND   m.name = pggen.arg('name')
AND   m.provider = pggen.arg('provider')
;

-- name: UploadModuleVersion :one
UPDATE module_versions
SET
    tarball = pggen.arg('tarball'),
    updated_at = pggen.arg('updated_at')
WHERE module_id = pggen.arg('module_id')
AND   version = pggen.arg('version')
RETURNING version;

-- name: DownloadModuleVersion :one
SELECT tarball
FROM module_versions
WHERE module_id = pggen.arg('module_id')
AND   version = pggen.arg('version')
;

-- name: DeleteModuleByID :one
DELETE
FROM modules
WHERE module_id = pggen.arg('id')
RETURNING module_id;
