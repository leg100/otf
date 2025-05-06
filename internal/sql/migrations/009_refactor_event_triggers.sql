CREATE OR REPLACE FUNCTION runs_notify_event() RETURNS trigger
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
                      'record', to_json(record));
    PERFORM pg_notify('events', notification::text);
    RETURN NULL;
END;
$$;

---- create above / drop below ----
CREATE OR REPLACE FUNCTION runs_notify_event() RETURNS trigger
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
                      'id', record.run_id);
    PERFORM pg_notify('events', notification::text);
    RETURN NULL;
END;
$$;
