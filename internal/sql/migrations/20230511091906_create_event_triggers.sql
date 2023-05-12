-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION notify_table_update() RETURNS TRIGGER
AS $notifiy_table_update$
    DECLARE
        record RECORD;
    BEGIN
        CASE TG_OP
            WHEN 'INSERT', 'UPDATE' THEN
                record = NEW;
            WHEN 'DELETE' THEN
                record = OLD;
            ELSE
                RETURN NULL;
        END CASE;

        PERFORM pg_notify(
            'events',
            json_build_object(
                'table', TG_TABLE_NAME,
                'op', TG_OP,
                'record', row_to_json(record)
            )::text
        );
        RETURN NULL;
    END;
$notifiy_table_update$ LANGUAGE plpgsql;

CREATE TRIGGER organization_events
    AFTER INSERT OR UPDATE OR DELETE
    ON organizations
    FOR EACH ROW
    EXECUTE FUNCTION notify_table_update();

CREATE TRIGGER workspace_events
    AFTER INSERT OR UPDATE OR DELETE
    ON workspaces
    FOR EACH ROW
    EXECUTE FUNCTION notify_table_update();

CREATE TRIGGER run_events
    AFTER INSERT OR UPDATE OR DELETE
    ON runs
    FOR EACH ROW
    EXECUTE FUNCTION notify_table_update();
-- +goose StatementEnd

-- +goose Down
DROP TRIGGER IF EXISTS run_events ON runs;
DROP TRIGGER IF EXISTS workspace_events ON workspaces;
DROP TRIGGER IF EXISTS organization_events ON organizations;
DROP FUNCTION notify_table_update;
