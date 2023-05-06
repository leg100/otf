-- +goose Up
ALTER TABLE workspaces ADD CONSTRAINT workspace_name_uniq UNIQUE (organization_name, name);

-- +goose Down
ALTER TABLE workspaces DROP CONSTRAINT workspace_name_uniq;
