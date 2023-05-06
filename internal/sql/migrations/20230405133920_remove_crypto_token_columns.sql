-- +goose Up
ALTER TABLE agent_tokens DROP COLUMN token;
ALTER TABLE tokens DROP COLUMN token;

-- +goose Down
ALTER TABLE tokens ADD COLUMN token TEXT;
ALTER TABLE agent_tokens ADD COLUMN token TEXT;
