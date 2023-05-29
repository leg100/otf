-- +goose Up
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION organizations_notify_event() RETURNS TRIGGER AS $$
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
                      'id', record.organization_id);
    PERFORM pg_notify('events', notification::text);
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION workspaces_notify_event() RETURNS TRIGGER AS $$
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
                      'id', record.workspace_id);
    PERFORM pg_notify('events', notification::text);
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION runs_notify_event() RETURNS TRIGGER AS $$
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
                      'id', record.run_id);
    PERFORM pg_notify('events', notification::text);
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION logs_notify_event() RETURNS TRIGGER AS $$
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
                      'id', record.chunk_id::text);
    PERFORM pg_notify('events', notification::text);
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION notification_configurations_notify_event() RETURNS TRIGGER AS $$
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
                      'id', record.notification_configuration_id);
    PERFORM pg_notify('events', notification::text);
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER notify_event
AFTER INSERT OR UPDATE OR DELETE ON organizations
    FOR EACH ROW EXECUTE PROCEDURE organizations_notify_event();

CREATE TRIGGER notify_event
AFTER INSERT OR UPDATE OR DELETE ON workspaces
    FOR EACH ROW EXECUTE PROCEDURE workspaces_notify_event();

CREATE TRIGGER notify_event
AFTER INSERT OR UPDATE OR DELETE ON runs
    FOR EACH ROW EXECUTE PROCEDURE runs_notify_event();

CREATE TRIGGER notify_event
AFTER INSERT OR UPDATE OR DELETE ON logs
    FOR EACH ROW EXECUTE PROCEDURE logs_notify_event();

CREATE TRIGGER notify_event
AFTER INSERT OR UPDATE OR DELETE ON notification_configurations
    FOR EACH ROW EXECUTE PROCEDURE notification_configurations_notify_event();
-- +goose StatementEnd

-- +goose Down
DROP TRIGGER IF EXISTS notify_event ON notification_configurations;
DROP TRIGGER IF EXISTS notify_event ON logs;
DROP TRIGGER IF EXISTS notify_event ON runs;
DROP TRIGGER IF EXISTS notify_event ON workspaces;
DROP TRIGGER IF EXISTS notify_event ON organizations;
DROP FUNCTION IF EXISTS notify_event;
