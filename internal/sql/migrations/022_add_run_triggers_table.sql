CREATE TABLE run_triggers (
    run_trigger_id text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    workspace_id text NOT NULL,
    sourceable_workspace_id text NOT NULL,
	FOREIGN KEY (workspace_id) REFERENCES workspaces(workspace_id) ON UPDATE CASCADE ON DELETE CASCADE,
	FOREIGN KEY (sourceable_workspace_id) REFERENCES workspaces(workspace_id) ON UPDATE CASCADE ON DELETE CASCADE
);

ALTER TABLE runs
    ADD COLUMN triggering_run_id TEXT REFERENCES runs(run_id) ON DELETE SET NULL;

---- create above / drop below ----
ALTER TABLE runs
	DROP COLUMN triggering_run_id;

DROP TABLE run_triggers;
