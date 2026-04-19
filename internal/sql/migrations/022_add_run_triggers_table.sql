-- Add support for run triggers.

CREATE TABLE run_triggers (
    run_trigger_id text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    workspace_id text NOT NULL,
    triggering_workspace_id text NOT NULL,
	FOREIGN KEY (workspace_id) REFERENCES workspaces(workspace_id) ON UPDATE CASCADE ON DELETE CASCADE,
	FOREIGN KEY (triggering_workspace_id) REFERENCES workspaces(workspace_id) ON UPDATE CASCADE ON DELETE CASCADE
);

ALTER TABLE runs
    ADD COLUMN triggering_run_id TEXT REFERENCES runs(run_id) ON DELETE SET NULL;

-- Add auto-apply run trigger column on workspaces table. Set to false for all
-- existing workspaces and set not null.
ALTER TABLE workspaces
    ADD COLUMN auto_apply_run_trigger BOOLEAN;

UPDATE workspaces
SET auto_apply_run_trigger = false;

ALTER TABLE workspaces
	ALTER COLUMN auto_apply_run_trigger SET NOT NULL;

---- create above / drop below ----
ALTER TABLE workspaces
	DROP COLUMN auto_apply_run_trigger;

ALTER TABLE runs
	DROP COLUMN triggering_run_id;

DROP TABLE run_triggers;
