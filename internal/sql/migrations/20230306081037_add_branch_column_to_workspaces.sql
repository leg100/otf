-- +goose Up
ALTER TABLE workspaces ADD COLUMN branch TEXT DEFAULT '';
ALTER TABLE workspaces
    ALTER COLUMN branch SET NOT NULL,
    ALTER COLUMN branch DROP DEFAULT
;

-- +goose Down
ALTER TABLE workspaces DROP COLUMN branch;
