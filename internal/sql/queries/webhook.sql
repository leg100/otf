-- name: InsertWebhook :one
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
)
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

-- name: FindWebhookByIDForUpdate :one
SELECT *
FROM webhooks
WHERE webhook_id = pggen.arg('webhook_id')
FOR UPDATE;

-- name: FindWebhookByRepoForUpdate :one
SELECT *
FROM webhooks
WHERE identifier = pggen.arg('identifier')
AND   cloud = pggen.arg('cloud')
FOR UPDATE;

-- name: FindWebhookByRepo :many
SELECT *
FROM webhooks
WHERE identifier = pggen.arg('identifier')
AND   cloud = pggen.arg('cloud');

-- name: DeleteWebhookByID :one
DELETE
FROM webhooks
WHERE webhook_id = pggen.arg('webhook_id')
RETURNING *
;
