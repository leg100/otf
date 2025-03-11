-- Trigger a workspace event whenever a run belonging to the workspace is
-- updated, so that downstream systems are informed when a workspace's current
-- run status changes.

CREATE FUNCTION workspace_run_notify_event() RETURNS trigger
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
                      'table','workspaces',
                      'action', 'UPDATE',
                      'id', record.workspace_id);
    PERFORM pg_notify('events', notification::text);
    RETURN NULL;
END;
$$;

DO $$BEGIN
    CREATE TRIGGER notify_workspace_event AFTER UPDATE ON runs FOR EACH ROW EXECUTE FUNCTION workspace_run_notify_event();
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

---- create above / drop below ----

DROP TRIGGER notify_workspace_event ON runs;
DROP FUNCTION workspace_run_notify_event;
