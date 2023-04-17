-- +goose Up
ALTER TABLE workspaces DROP CONSTRAINT lock_run_id_fk;
ALTER TABLE workspaces ADD CONSTRAINT lock_run_id_fk
    FOREIGN KEY (lock_run_id) REFERENCES runs(run_id) ON UPDATE CASCADE;

-- +goose Down
ALTER TABLE workspaces DROP CONSTRAINT lock_run_id_fk;
ALTER TABLE workspaces ADD CONSTRAINT lock_run_id_fk
    FOREIGN KEY (lock_run_id) REFERENCES runs(run_id) ON UPDATE CASCADE ON DELETE CASCADE;

