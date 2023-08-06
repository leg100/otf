-- +goose Up
ALTER TABLE workspaces ADD COLUMN allow_cli_apply BOOL NOT NULL DEFAULT false;

-- +goose Down
ALTER TABLE workspaces DROP COLUMN allow_cli_apply;
