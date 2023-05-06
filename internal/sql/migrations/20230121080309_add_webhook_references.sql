-- add column counting the number of workspaces and modules that
-- are 'connected' to the webhook
--
-- +goose Up
ALTER TABLE webhooks ADD COLUMN connected INTEGER DEFAULT 0;

-- account for connected workspaces
UPDATE webhooks h
SET connected = connected + connections.total
FROM (
    SELECT webhook_id, count(*) AS total
    FROM workspace_repos
    GROUP BY webhook_id
) AS connections
WHERE connections.webhook_id = h.webhook_id
;

-- account for connected modules
UPDATE webhooks h
SET connected = connected + connections.total
FROM (
    SELECT webhook_id, count(*) AS total
    FROM module_repos
    GROUP BY webhook_id
) AS connections
WHERE connections.webhook_id = h.webhook_id
;

-- remove webhooks with no connections
DELETE
FROM webhooks
WHERE connected = 0
;

-- connections all accounted for, we can now enforce not null
-- constraint and remove default
ALTER TABLE webhooks
    ALTER COLUMN connected SET NOT NULL,
    ALTER COLUMN connected DROP DEFAULT
;

-- +goose Down
ALTER TABLE webhooks DROP COLUMN connected;
