-- name: InsertWebhook :one
WITH inserted AS (
    INSERT INTO webhooks (
        webhook_id,
        vcs_id,
        vcs_provider_id,
        secret,
        identifier
    ) VALUES (
        pggen.arg('webhook_id'),
        pggen.arg('vcs_id'),
        pggen.arg('vcs_provider_id'),
        pggen.arg('secret'),
        pggen.arg('identifier')
    )
    RETURNING *
)
SELECT
    w.webhook_id,
    w.vcs_id,
    w.vcs_provider_id,
    w.secret,
    w.identifier,
    v.cloud
FROM inserted w
JOIN vcs_providers v USING (vcs_provider_id);

-- name: UpdateWebhookVCSID :one
UPDATE webhooks
SET vcs_id = pggen.arg('vcs_id')
WHERE webhook_id = pggen.arg('webhook_id')
RETURNING *;

-- name: FindWebhooks :many
SELECT
    w.webhook_id,
    w.vcs_id,
    w.vcs_provider_id,
    w.secret,
    w.identifier,
    v.cloud
FROM webhooks w
JOIN vcs_providers v USING (vcs_provider_id);

-- name: FindWebhookByID :one
SELECT
    w.webhook_id,
    w.vcs_id,
    w.vcs_provider_id,
    w.secret,
    w.identifier,
    v.cloud
FROM webhooks w
JOIN vcs_providers v USING (vcs_provider_id)
WHERE w.webhook_id = pggen.arg('webhook_id');

-- name: FindWebhookByRepoAndProvider :many
SELECT
    w.webhook_id,
    w.vcs_id,
    w.vcs_provider_id,
    w.secret,
    w.identifier,
    v.cloud
FROM webhooks w
JOIN vcs_providers v USING (vcs_provider_id)
WHERE identifier = pggen.arg('identifier')
AND   vcs_provider_id = pggen.arg('vcs_provider_id');

-- name: FindWebhooksByProvider :many
SELECT
    w.webhook_id,
    w.vcs_id,
    w.vcs_provider_id,
    w.secret,
    w.identifier,
    v.cloud
FROM webhooks w
JOIN vcs_providers v USING (vcs_provider_id)
WHERE w.vcs_provider_id = pggen.arg('vcs_provider_id');

-- name: FindUnreferencedWebhooks :many
SELECT
    w.webhook_id,
    w.vcs_id,
    w.vcs_provider_id,
    w.secret,
    w.identifier,
    v.cloud
FROM webhooks w
JOIN vcs_providers v USING (vcs_provider_id)
WHERE NOT EXISTS (
    SELECT FROM repo_connections c
    WHERE c.webhook_id = w.webhook_id
);

-- name: DeleteWebhookByID :one
DELETE
FROM webhooks
WHERE webhook_id = pggen.arg('webhook_id')
RETURNING *;
