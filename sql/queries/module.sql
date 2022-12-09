-- name: InsertModule :exec
INSERT INTO modules (
    module_id,
    created_at,
    updated_at,
    name,
    provider,
    organization
) VALUES (
    pggen.arg('id'),
    pggen.arg('created_at'),
    pggen.arg('updated_at'),
    pggen.arg('name'),
    pggen.arg('provider'),
    pggen.arg('organization')
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
    pggen.arg('module')
)
RETURNING *;

-- name ListModulesByOrganization :many
SELECT
    module_id,
    created_at,
    updated_at,
    name,
    provider,
    (o.*)::"organizations" AS organization
FROM modules
JOIN organizations o USING (organization_id)
WHERE o.name = pggen.arg('organization_name')
;

-- name: FindModuleByName :one
SELECT
    module_id,
    created_at,
    updated_at,
    name,
    provider,
    (o.*)::"organizations" AS organization
FROM modules
JOIN organizations o USING (organization_id)
WHERE name = pggen.arg('name')
AND   provider = pggen.arg('provider')
;

-- name: UploadModuleVersion :one
UPDATE module_versions (
    tarball,
    updated_at
) VALUES (
    pggen.arg('tarball'),
    pggen.arg('updated_at')
)
RETURNING version;

-- name: DownloadModuleVersion :one
SELECT tarball
FROM modules
WHERE name = pggen.arg('name')
AND   provider = pggen.arg('provider')
AND   version = pggen.arg('version')
;

-- name: DeleteModuleByID :one
DELETE
FROM modules
WHERE module_id = pggen.arg('id')
RETURNING module_id;
