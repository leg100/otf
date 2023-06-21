-- +goose Up
ALTER TABLE state_version_outputs ALTER COLUMN value TYPE bytea USING value::bytea;

-- +goose Down
ALTER TABLE state_version_outputs ALTER COLUMN value TYPE text;
