ALTER TABLE policy_sets
    ADD COLUMN kind text NOT NULL DEFAULT 'sentinel',
    ADD COLUMN engine_version text NOT NULL DEFAULT 'latest';

UPDATE policy_sets
SET kind = 'sentinel'
WHERE kind = '';

UPDATE policy_sets
SET engine_version = 'latest'
WHERE engine_version = '';

---- create above / drop below ----

ALTER TABLE policy_sets
    DROP COLUMN engine_version,
    DROP COLUMN kind;
