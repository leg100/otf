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

-- name: InsertModuleRepo :exec
INSERT INTO module_repos (
    webhook_id,
    vcs_provider_id,
    module_id
) VALUES (
    pggen.arg('webhook_id'),
    pggen.arg('vcs_provider_id'),
    pggen.arg('module_id')
);

-- name: InsertModuleVersion :one
INSERT INTO module_versions (
    module_version_id,
    version,
    created_at,
    updated_at,
    module_id
) VALUES (
    pggen.arg('module_version_id'),
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
    (o.*)::"organizations" AS organization,
    (r.*)::"module_repos" AS module_repo,
    (h.*)::"webhooks" AS webhook,
    (
        SELECT array_agg(v.*) AS versions
        FROM module_versions v
        WHERE v.module_id = m.module_id
    ) AS versions
FROM modules m
JOIN organizations o USING (organization_id)
LEFT JOIN (module_repos r JOIN webhooks h USING (webhook_id)) USING (module_id)
WHERE o.name = pggen.arg('organization_name')
;

-- name: FindModuleByName :one
SELECT
    m.module_id,
    m.created_at,
    m.updated_at,
    m.name,
    m.provider,
    (o.*)::"organizations" AS organization,
    (r.*)::"module_repos" AS module_repo,
    (h.*)::"webhooks" AS webhook,
    (
        SELECT array_agg(v.*) AS versions
        FROM module_versions v
        WHERE v.module_id = m.module_id
    ) AS versions
FROM modules m
JOIN organizations o USING (organization_id)
LEFT JOIN (module_repos r JOIN webhooks h USING (webhook_id)) USING (module_id)
WHERE o.name = pggen.arg('organizaton_name')
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
    (o.*)::"organizations" AS organization,
    (r.*)::"module_repos" AS module_repo,
    (h.*)::"webhooks" AS webhook,
    (
        SELECT array_agg(v.*) AS versions
        FROM module_versions v
        WHERE v.module_id = m.module_id
    ) AS versions
FROM modules m
JOIN organizations o USING (organization_id)
LEFT JOIN (module_repos r JOIN webhooks h USING (webhook_id)) USING (module_id)
WHERE m.module_id = pggen.arg('id')
;

-- name: FindModuleByWebhookID :one
SELECT
    m.module_id,
    m.created_at,
    m.updated_at,
    m.name,
    m.provider,
    (o.*)::"organizations" AS organization,
    (r.*)::"module_repos" AS module_repo,
    (h.*)::"webhooks" AS webhook,
    (
        SELECT array_agg(v.*) AS versions
        FROM module_versions v
        WHERE v.module_id = m.module_id
    ) AS versions
FROM modules m
JOIN organizations o USING (organization_id)
JOIN (module_repos r JOIN webhooks h USING (webhook_id)) USING (module_id)
WHERE h.webhook_id = pggen.arg('webhook_id')
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

-- name: DeleteModuleByID :one
DELETE
FROM modules
WHERE module_id = pggen.arg('id')
RETURNING module_id;
