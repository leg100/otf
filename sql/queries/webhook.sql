-- FindOrInsertWebhook idempotently inserts a webhook,
-- returning it if it already exists.
--
-- name: InsertWebhook :exec
INSERT INTO webhooks (
    webhook_id,
    vcs_id,
    secret,
    identifier,
    cloud
) VALUES (
    pggen.arg('webhook_id'),
    pggen.arg('vcs_id'),
    pggen.arg('secret'),
    pggen.arg('identifier'),
    pggen.arg('cloud')
);

-- name: UpdateWebhookVCSID :exec
UPDATE webhooks
SET vcs_id = pggen.arg('vcs_id')
WHERE webhook_id = pggen.arg('webhook_id');

-- name: FindWebhookByID :one
SELECT *
FROM webhooks
WHERE webhook_id = pggen.arg('webhook_id');

-- name: FindWebhookByRepo :one
SELECT *
FROM webhooks
WHERE identifier = pggen.arg('identifier')
AND   cloud = pggen.arg('cloud');

-- name: DeleteWebhook :exec
DELETE
FROM webhooks
WHERE webhook_id = pggen.arg('webhook_id');
