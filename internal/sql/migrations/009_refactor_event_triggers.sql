CREATE TABLE event_actions (
	action TEXT NOT NULL,
);

INSERT INTO event_actions (action) VALUES (
	'INSERT',
	'UPDATE',
	'DELETE'
);

CREATE TABLE events (
	id SERIAL PRIMARY KEY,
	action TEXT NOT NULL,
	table TEXT NOT NULL,
	payload BYTEA NOT NULL,
	FOREIGN KEY (action) REFERENCES event_actions(action)
);

CREATE FUNCTION events_notify_event() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
-- DECLARE
BEGIN
    PERFORM pg_notify('events', ''::text);
    RETURN NULL;
END;
$$;

CREATE TRIGGER notify_event AFTER INSERT ON events FOR EACH ROW EXECUTE FUNCTION events_notify_event();

---- create above / drop below ----
DROP TRIGGER notify_event ON events;
DROP FUNCTION events_notify_event;
DROP TABLE events;
DROP TABLE event_actions;
