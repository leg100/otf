-- FindOrInsertWebhook idempotently inserts a webhook,
-- returning it if it already exists.
--
-- name: FindOrInsertWebhook :one
INSERT INTO webhooks (
    webhook_id,
    endpoint,
    secret,
    identifier,
    http_url
) VALUES (
    pggen.arg('webhook_id'),
    pggen.arg('endpoint'),
    pggen.arg('secret'),
    pggen.arg('identifier'),
    pggen.arg('http_url')
) ON CONFLICT DO NOTHING
RETURNING *;

-- name: UpdateWebhookVCSID :exec
UPDATE webhooks
SET vcs_id = pggen.arg('vcs_id')
WHERE webhook_id = pggen.arg('webhook_id');

-- name: FindWebhookSecret :one
SELECT secret
FROM webhooks
WHERE webhook_id = pggen.arg('webhook_id');

-- name: DeleteWebhook :exec
DELETE
FROM webhooks
WHERE webhook_id = pggen.arg('webhook_id');
