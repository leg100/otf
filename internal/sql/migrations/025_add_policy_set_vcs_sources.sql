ALTER TABLE policy_sets
    ADD COLUMN source text NOT NULL DEFAULT 'manual',
    ADD COLUMN vcs_provider_id text REFERENCES vcs_providers(vcs_provider_id) ON UPDATE CASCADE ON DELETE SET NULL,
    ADD COLUMN vcs_repo text,
    ADD COLUMN vcs_ref text NOT NULL DEFAULT '',
    ADD COLUMN vcs_path text NOT NULL DEFAULT '',
    ADD COLUMN vcs_policy_paths text[] NOT NULL DEFAULT '{}',
    ADD COLUMN last_synced_at timestamp with time zone;

ALTER TABLE policies
    ADD COLUMN path text NOT NULL DEFAULT '';

UPDATE policy_sets
SET source = 'manual'
WHERE source = '';

---- create above / drop below ----

ALTER TABLE policies
    DROP COLUMN path;

ALTER TABLE policy_sets
    DROP COLUMN last_synced_at,
    DROP COLUMN vcs_policy_paths,
    DROP COLUMN vcs_path,
    DROP COLUMN vcs_ref,
    DROP COLUMN vcs_repo,
    DROP COLUMN vcs_provider_id,
    DROP COLUMN source;
