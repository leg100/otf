CREATE TABLE run_triggers (
    run_trigger_id text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    workspace_id text NOT NULL,
    sourceable_workspace_id text NOT NULL,
	FOREIGN KEY (workspace_id) REFERENCES workspaces(workspace_id) ON UPDATE CASCADE ON DELETE CASCADE,
	FOREIGN KEY (sourceable_workspace_id) REFERENCES workspaces(workspace_id) ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE OR REPLACE FUNCTION run_triggers_notify_event() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
DECLARE
    record RECORD;
    notification JSON;
BEGIN
    IF (TG_OP = 'DELETE') THEN
        record = OLD;
    ELSE
        record = NEW;
    END IF;
    notification = json_build_object(
                      'table',TG_TABLE_NAME,
                      'action', TG_OP,
                      'id', record.run_trigger_id);
    PERFORM pg_notify('events', notification::text);
    RETURN NULL;
END;
$$;

CREATE OR REPLACE TRIGGER notify_event AFTER INSERT OR DELETE OR UPDATE ON run_triggers FOR EACH ROW EXECUTE FUNCTION runtriggers_notify_event();

---- create above / drop below ----

DROP TRIGGER run_triggers_notify_event;
DROP TABLE run_triggers;
