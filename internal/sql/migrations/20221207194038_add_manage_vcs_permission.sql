-- +goose Up
ALTER TABLE teams ADD COLUMN permission_manage_vcs BOOL NOT NULL DEFAULT FALSE;


-- +goose Down
ALTER TABLE teams DROP COLUMN permission_manage_vcs;
