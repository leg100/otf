-- +goose Up
ALTER TABLE webhooks DROP COLUMN http_url;
ALTER TABLE webhooks ALTER COLUMN vcs_id SET NOT NULL;
ALTER TABLE webhooks ADD CONSTRAINT webhooks_cloud_id_uniq UNIQUE (cloud, identifier);

-- +goose Down
ALTER TABLE webhooks DROP CONSTRAINT webhooks_cloud_id_uniq;
ALTER TABLE webhooks ALTER COLUMN vcs_id DROP NOT NULL;
ALTER TABLE webhooks ADD COLUMN http_url TEXT;
