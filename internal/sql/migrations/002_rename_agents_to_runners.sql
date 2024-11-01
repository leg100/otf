-- rename agents table to runners
ALTER TABLE agents RENAME TO runners;
ALTER TABLE runners RENAME COLUMN agent_id TO runner_id;
ALTER TABLE agent_statuses RENAME TO runner_statuses;
ALTER TABLE jobs RENAME COLUMN agent_id TO runner_id;

-- rename trigger and trigger function
DROP TRIGGER notify_event ON runners;
DROP FUNCTION agents_notify_event();
CREATE FUNCTION runners_notify_event() RETURNS trigger
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
                      'id', record.runner_id);
    PERFORM pg_notify('events', notification::text);
    RETURN NULL;
END;
$$;
CREATE TRIGGER notify_event AFTER INSERT OR DELETE OR UPDATE ON runners FOR EACH ROW EXECUTE FUNCTION public.runners_notify_event();


---- create above / drop below ----

ALTER TABLE runners RENAME TO agents;
ALTER TABLE agents RENAME COLUMN runner_id TO agent_id;
ALTER TABLE runner_statuses RENAME TO agent_statuses;
ALTER TABLE jobs RENAME COLUMN runner_id TO agent_id;

DROP TRIGGER notify_event ON agents;
DROP FUNCTION runners_notify_event();
CREATE FUNCTION agents_notify_event() RETURNS trigger
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
                      'id', record.agent_id);
    PERFORM pg_notify('events', notification::text);
    RETURN NULL;
END;
$$;
CREATE TRIGGER notify_event AFTER INSERT OR DELETE OR UPDATE ON agents FOR EACH ROW EXECUTE FUNCTION public.agents_notify_event();
