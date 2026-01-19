-- Add executor column to runners table. Add foreign key to new executors
-- table, which is populated with possible executor kinds. Set default to
-- 'fork' executor kind.

CREATE TABLE executors (
    executor text PRIMARY KEY
);
INSERT INTO executors VALUES ('fork'), ('kubernetes');

ALTER TABLE runners
	ADD COLUMN executor text,
    ADD CONSTRAINT runners_executor_fkey FOREIGN KEY (executor) REFERENCES executors(executor) ON UPDATE CASCADE;

UPDATE runners
SET executor = 'fork';

ALTER TABLE runners ALTER COLUMN executor SET NOT NULL;

---- create above / drop below ----
ALTER TABLE runners DROP COLUMN executor;
DROP TABLE executors;
