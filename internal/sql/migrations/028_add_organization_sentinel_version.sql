ALTER TABLE organizations
    ADD COLUMN sentinel_version text NOT NULL DEFAULT 'latest';

UPDATE organizations
SET sentinel_version = 'latest'
WHERE sentinel_version = '';

---- create above / drop below ----

ALTER TABLE organizations
    DROP COLUMN sentinel_version;
