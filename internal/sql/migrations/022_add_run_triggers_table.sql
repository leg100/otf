CREATE TABLE run_triggers (
    run_trigger_id text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    workspace_id text NOT NULL,
    sourceable_workspace_id text NOT NULL,
	FOREIGN KEY (workspace_id) REFERENCES workspaces(workspace_id) ON UPDATE CASCADE ON DELETE CASCADE,
	FOREIGN KEY (sourceable_workspace_id) REFERENCES workspaces(workspace_id) ON UPDATE CASCADE ON DELETE CASCADE
);
---- create above / drop below ----
DROP TABLE run_triggers;
