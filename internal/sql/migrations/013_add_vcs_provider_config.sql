ALTER TABLE vcs_providers ADD COLUMN config JSON;
---- create above / drop below ----
ALTER TABLE vcs_providers DROP COLUMN config;
