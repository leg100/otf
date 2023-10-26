-- +goose Up
CREATE TABLE IF NOT EXISTS state_version_statuses (
    status TEXT PRIMARY KEY
);

INSERT INTO state_version_statuses (status) VALUES
	('pending'),
	('finalized'),
	('discarded');

ALTER TABLE state_versions
    ALTER COLUMN state DROP NOT NULL,
    ADD COLUMN status TEXT,
    ADD CONSTRAINT status_fk FOREIGN KEY (status)
        REFERENCES state_version_statuses ON UPDATE CASCADE;

UPDATE state_versions
SET status = 'finalized';

ALTER TABLE state_versions
    ALTER COLUMN status SET NOT NULL;

-- +goose Down
ALTER TABLE state_versions
    ALTER COLUMN state SET NOT NULL,
    DROP COLUMN status;

DROP TABLE IF EXISTS state_version_statuses;
