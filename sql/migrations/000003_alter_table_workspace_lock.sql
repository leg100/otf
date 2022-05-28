-- +goose Up
ALTER TABLE workspaces DROP COLUMN locked;
ALTER TABLE workspaces ADD COLUMN lock_run_id TEXT;
ALTER TABLE workspaces ADD COLUMN lock_user_id TEXT;
ALTER TABLE workspaces ADD CONSTRAINT lock_run_id_fk FOREIGN KEY (lock_run_id) REFERENCES runs (run_id) ON UPDATE CASCADE ON DELETE CASCADE;
ALTER TABLE workspaces ADD CONSTRAINT lock_user_id_fk FOREIGN KEY (lock_user_id) REFERENCES users (user_id) ON UPDATE CASCADE ON DELETE CASCADE;
ALTER TABLE workspaces ADD CONSTRAINT lock_check CHECK (lock_user_id IS NULL OR lock_run_id IS NULL);

-- +goose Down
ALTER TABLE workspaces DROP CONSTRAINT lock_check;
ALTER TABLE workspaces DROP CONSTRAINT lock_user_id_fk;
ALTER TABLE workspaces DROP CONSTRAINT lock_run_id_fk;
ALTER TABLE workspaces DROP COLUMN lock_user_id;
ALTER TABLE workspaces DROP COLUMN lock_run_id;
ALTER TABLE workspaces ADD COLUMN locked BOOLEAN NOT NULL;
