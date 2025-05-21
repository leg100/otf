-- Introduce new table "logs" to supplement the existing "logs" table, which is
-- renamed to "chunks": the newly named chunks table stores chunks of logs for
-- incomplete runs; whereas the new logs table stores complete logs for
-- completed runs.
ALTER TABLE logs RENAME TO chunks;
CREATE TABLE logs (
    run_id text NOT NULL,
    phase text NOT NULL,
    logs bytea NOT NULL,
	CONSTRAINT run_id_phase PRIMARY KEY(run_id, phase),
	FOREIGN KEY (run_id) REFERENCES runs(run_id) ON UPDATE CASCADE ON DELETE CASCADE,
	FOREIGN KEY (phase) REFERENCES phases(phase) ON UPDATE CASCADE ON DELETE CASCADE
);
---- create above / drop below ----
DROP TABLE logs;
ALTER TABLE chunks RENAME TO logs;
