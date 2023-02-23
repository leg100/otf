-- +goose Up
ALTER TABLE state_version_outputs ALTER COLUMN value TYPE BYTEA USING value::TEXT::BYTEA;

-- +goose Down
ALTER TABLE state_version_outputs ALTER COLUMN value TYPE TEXT USING value::BYTEA::TEXT;
