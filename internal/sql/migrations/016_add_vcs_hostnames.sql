-- Add two fields to vcs_providers, one optional (http_url), one mandatory
-- (api_url). Populate the api_url field before setting not null.
ALTER TABLE vcs_providers
	ADD COLUMN http_url TEXT,
	ADD COLUMN base_url TEXT;

UPDATE vcs_providers
SET base_url = 'https://github.com/api/v3'
WHERE vcs_kind = 'github-app' OR vcs_kind = 'github-token';

UPDATE vcs_providers
SET base_url = 'https://gitlab.com/api/v4'
WHERE vcs_kind = 'gitlab';

UPDATE vcs_providers
SET base_url = 'https://next.forgejo.org'
WHERE vcs_kind = 'forgejo';

ALTER TABLE vcs_providers
ALTER COLUMN base_url SET NOT NULL;

---- create above / drop below ----
ALTER TABLE vcs_providers
	DROP COLUMN http_url,
	DROP COLUMN base_url;
