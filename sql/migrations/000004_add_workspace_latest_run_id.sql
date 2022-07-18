-- +goose Up
ALTER TABLE workspaces ADD COLUMN latest_run_id TEXT;
ALTER TABLE workspaces ADD CONSTRAINT latest_run_id_fk FOREIGN KEY (latest_run_id) REFERENCES runs ON UPDATE CASCADE;

-- +goose Down
ALTER TABLE workspaces DROP CONSTRAINT latest_run_id_fk;
ALTER TABLE workspaces DROP COLUMN latest_run_id;
