-- +goose Up
ALTER TABLE workspaces ADD COLUMN trigger_patterns TEXT[];

-- +goose Down
ALTER TABLE workspaces DROP COLUMN trigger_patterns;
