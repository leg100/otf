CREATE TABLE event_actions (
	action TEXT PRIMARY KEY
);

INSERT INTO event_actions (action) VALUES
	('INSERT'),
	('UPDATE'),
	('DELETE')
;

CREATE TABLE events (
	id SERIAL PRIMARY KEY,
	action TEXT NOT NULL,
	type TEXT NOT NULL,
	payload BYTEA NOT NULL,
	FOREIGN KEY (action) REFERENCES event_actions(action)
);

---- create above / drop below ----
DROP TABLE events;
DROP TABLE event_actions;
