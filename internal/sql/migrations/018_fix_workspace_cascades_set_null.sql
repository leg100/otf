-- Fix constraints on workspaces table: before this migration, a run with a
-- workspace lock or a run that is the latest run for a workspace could not be
-- deleted without first deleting its workspace; and deleting a user that has a
-- workspace lock would delete the workspace!
--
-- After this migration, deleting a run sets any referencing fields in its
-- workspace to NULL; and deleting a user that has a workspace lock will set
-- the referencing field to NULL. (Effectively deleting a user or run lock
-- unlocks the workspace).
ALTER TABLE workspaces DROP CONSTRAINT latest_run_id_fk;
ALTER TABLE workspaces ADD CONSTRAINT latest_run_id_fk
	FOREIGN KEY (latest_run_id)
	REFERENCES runs(run_id)
	ON UPDATE CASCADE
	ON DELETE SET NULL;
ALTER TABLE workspaces DROP CONSTRAINT lock_run_id_fk;
ALTER TABLE workspaces ADD CONSTRAINT lock_run_id_fk
	FOREIGN KEY (lock_run_id)
	REFERENCES runs(run_id)
	ON UPDATE CASCADE
	ON DELETE SET NULL;
ALTER TABLE workspaces DROP CONSTRAINT workspaces_lock_username_fkey;
ALTER TABLE workspaces ADD CONSTRAINT workspaces_lock_username_fkey
	FOREIGN KEY (lock_username)
	REFERENCES users(username)
	ON UPDATE CASCADE
	ON DELETE SET NULL;
---- create above / drop below ----
ALTER TABLE workspaces DROP CONSTRAINT latest_run_id_fk;
ALTER TABLE workspaces ADD CONSTRAINT latest_run_id_fk
	FOREIGN KEY (latest_run_id)
	REFERENCES runs(run_id)
	ON UPDATE CASCADE;
ALTER TABLE workspaces DROP CONSTRAINT lock_run_id_fk;
ALTER TABLE workspaces ADD CONSTRAINT lock_run_id_fk
	FOREIGN KEY (lock_run_id)
	REFERENCES runs(run_id)
	ON UPDATE CASCADE;
ALTER TABLE workspaces DROP CONSTRAINT workspaces_lock_username_fkey;
ALTER TABLE workspaces ADD CONSTRAINT workspaces_lock_username_fkey
	FOREIGN KEY (lock_username)
	REFERENCES users(username)
	ON DELETE CASCADE;
