ALTER TABLE vcs_providers DROP COLUMN github_app_id;
ALTER TABLE vcs_providers ADD COLUMN hostname TEXT;
---- create above / drop below ----
ALTER TABLE vcs_providers ADD COLUMN github_app_id BIGINT;
-- TODO: re-populate github_app_id
