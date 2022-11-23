-- FindOrInsertWebhook idempotently inserts a webhook,
-- returning it if it already exists.
--
-- name: FindOrInsertWebhook :one
INSERT INTO webhooks (
    webhook_id,
    secret,
    identifier,
    http_url
) VALUES (
    pggen.arg('webhook_id'),
    pggen.arg('secret'),
    pggen.arg('identifier'),
    pggen.arg('http_url')
) ON CONFLICT DO NOTHING
RETURNING *;

-- name: FindWebhookSecret :one
SELECT secret
FROM webhooks
WHERE webhook_id = pggen.arg('webhook_id');

-- name: DeleteWebhook :exec
DELETE
FROM webhooks
WHERE webhook_id = pggen.arg('webhook_id');
