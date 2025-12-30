CREATE INDEX IF NOT EXISTS configuration_versions_created_at_idx ON configuration_versions(created_at);
---- create above / drop below ----
DROP INDEX configuration_versions_created_at_idx;
