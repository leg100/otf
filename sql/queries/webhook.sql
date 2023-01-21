-- name: InsertWebhook :one
INSERT INTO webhooks (
    webhook_id,
    secret,
    identifier,
    cloud,
    connected
) VALUES (
    pggen.arg('webhook_id'),
    pggen.arg('secret'),
    pggen.arg('identifier'),
    pggen.arg('cloud'),
    1
) ON CONFLICT (identifier, cloud) DO
    UPDATE
    SET connected = webhooks.connected + 1
RETURNING *
;

-- name: UpdateWebhookVCSID :one
UPDATE webhooks
SET vcs_id = pggen.arg('vcs_id')
WHERE webhook_id = pggen.arg('webhook_id')
RETURNING *
;

-- name: FindWebhookByID :one
SELECT *
FROM webhooks
WHERE webhook_id = pggen.arg('webhook_id');

-- name: FindWebhookByRepo :one
SELECT *
FROM webhooks
WHERE identifier = pggen.arg('identifier')
AND   cloud = pggen.arg('cloud');

-- name: DisconnectWebhook :one
UPDATE webhooks
SET connected = connected - 1
WHERE webhook_id = pggen.arg('webhook_id')
RETURNING connected
;

-- name: DeleteWebhook :exec
DELETE
FROM webhooks
WHERE webhook_id = pggen.arg('webhook_id');
