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
    pggen.arg('id'),
    pggen.arg('created_at'),
    pggen.arg('updated_at'),
    pggen.arg('name'),
    pggen.arg('provider'),
    pggen.arg('status'),
    pggen.arg('organization_name')
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
    pggen.arg('module_version_id'),
    pggen.arg('version'),
    pggen.arg('created_at'),
    pggen.arg('updated_at'),
    pggen.arg('module_id'),
    pggen.arg('status')
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
    (r.*)::"repo_connections" AS module_connection,
    (h.*)::"webhooks" AS webhook,
    (
        SELECT array_agg(v.*) AS versions
        FROM module_versions v
        WHERE v.module_id = m.module_id
    ) AS versions
FROM modules m
LEFT JOIN (repo_connections r JOIN webhooks h USING (webhook_id)) USING (module_id)
WHERE m.organization_name = pggen.arg('organization_name')
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
    (r.*)::"repo_connections" AS module_connection,
    (h.*)::"webhooks" AS webhook,
    (
        SELECT array_agg(v.*) AS versions
        FROM module_versions v
        WHERE v.module_id = m.module_id
    ) AS versions
FROM modules m
LEFT JOIN (repo_connections r JOIN webhooks h USING (webhook_id)) USING (module_id)
WHERE m.organization_name = pggen.arg('organization_name')
AND   m.name = pggen.arg('name')
AND   m.provider = pggen.arg('provider')
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
    (r.*)::"repo_connections" AS module_connection,
    (h.*)::"webhooks" AS webhook,
    (
        SELECT array_agg(v.*) AS versions
        FROM module_versions v
        WHERE v.module_id = m.module_id
    ) AS versions
FROM modules m
LEFT JOIN (repo_connections r JOIN webhooks h USING (webhook_id)) USING (module_id)
WHERE m.module_id = pggen.arg('id')
;

-- name: FindModuleByWebhookID :one
SELECT
    m.module_id,
    m.created_at,
    m.updated_at,
    m.name,
    m.provider,
    m.status,
    m.organization_name,
    (r.*)::"repo_connections" AS module_connection,
    (h.*)::"webhooks" AS webhook,
    (
        SELECT array_agg(v.*) AS versions
        FROM module_versions v
        WHERE v.module_id = m.module_id
    ) AS versions
FROM modules m
JOIN (repo_connections r JOIN webhooks h USING (webhook_id)) USING (module_id)
WHERE h.webhook_id = pggen.arg('webhook_id')
;

-- name: UpdateModuleStatusByID :one
UPDATE modules
SET status = pggen.arg('status')
WHERE module_id = pggen.arg('module_id')
RETURNING module_id
;

-- name: InsertModuleTarball :one
INSERT INTO module_tarballs (
    tarball,
    module_version_id
) VALUES (
    pggen.arg('tarball'),
    pggen.arg('module_version_id')
)
RETURNING module_version_id;

-- name: FindModuleTarball :one
SELECT tarball
FROM module_tarballs
WHERE module_version_id = pggen.arg('module_version_id')
;

-- name: UpdateModuleVersionStatusByID :one
UPDATE module_versions
SET
    status = pggen.arg('status'),
    status_error = pggen.arg('status_error')
WHERE module_version_id = pggen.arg('module_version_id')
RETURNING *
;
