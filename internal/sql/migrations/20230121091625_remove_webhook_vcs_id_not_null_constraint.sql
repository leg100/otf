-- +goose Up
ALTER TABLE webhooks ALTER COLUMN vcs_id DROP NOT NULL;

-- +goose Down
ALTER TABLE webhooks ALTER COLUMN vcs_id SET NOT NULL;
