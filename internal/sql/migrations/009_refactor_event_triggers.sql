-- Migrate from sending events containing only the id of a changed row, to instead sending the JSON encoded record of the row.
CREATE OR REPLACE FUNCTION build_and_send_event() RETURNS trigger
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
                      'record', record);
    PERFORM pg_notify('events', notification::text);
    RETURN NULL;
END;
$$;

ALTER TABLE runs ADD COLUMN updated_at TIMESTAMP WITH TIME ZONE;
UPDATE runs
SET updated_at = runs.created_at;

CREATE OR REPLACE TRIGGER notify_event AFTER INSERT OR DELETE OR UPDATE ON agent_pools FOR EACH ROW EXECUTE FUNCTION build_and_send_event();
CREATE OR REPLACE TRIGGER notify_event AFTER INSERT OR DELETE OR UPDATE ON runners FOR EACH ROW EXECUTE FUNCTION build_and_send_event();
CREATE OR REPLACE TRIGGER notify_event AFTER INSERT OR DELETE OR UPDATE ON jobs FOR EACH ROW EXECUTE FUNCTION build_and_send_event();
CREATE OR REPLACE TRIGGER notify_event AFTER INSERT OR DELETE OR UPDATE ON notification_configurations FOR EACH ROW EXECUTE FUNCTION build_and_send_event();
CREATE OR REPLACE TRIGGER notify_event AFTER INSERT OR DELETE OR UPDATE ON logs FOR EACH ROW EXECUTE FUNCTION build_and_send_event();
CREATE OR REPLACE TRIGGER notify_event AFTER INSERT OR DELETE OR UPDATE ON organizations FOR EACH ROW EXECUTE FUNCTION build_and_send_event();
CREATE OR REPLACE TRIGGER notify_event AFTER INSERT OR DELETE OR UPDATE ON workspaces FOR EACH ROW EXECUTE FUNCTION build_and_send_event();
CREATE OR REPLACE TRIGGER notify_event AFTER INSERT OR DELETE OR UPDATE ON runs FOR EACH ROW EXECUTE FUNCTION build_and_send_event();

DROP FUNCTION agent_pools_notify_event;
DROP FUNCTION runners_notify_event;
DROP FUNCTION jobs_notify_event;
DROP FUNCTION notification_configurations_notify_event;
DROP FUNCTION logs_notify_event;
DROP FUNCTION organizations_notify_event;
DROP FUNCTION workspaces_notify_event;
DROP FUNCTION runs_notify_event;
DROP TRIGGER notify_workspace_event ON runs;
DROP FUNCTION workspace_run_notify_event;
---- create above / drop below ----
