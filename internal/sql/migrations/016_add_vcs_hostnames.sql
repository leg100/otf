-- Add two fields to vcs_providers, one optional (http_url), one mandatory
-- (api_url). Populate the api_url field before setting not null.
ALTER TABLE vcs_providers
	ADD COLUMN http_url TEXT,
	ADD COLUMN api_url TEXT;

UPDATE vcs_providers
SET api_url = 'https://github.com/api/v3'
WHERE vcs_kind = 'github-app' OR vcs_kind = 'github-token';

UPDATE vcs_providers
SET api_url = 'https://gitlab.com/api/v4'
WHERE vcs_kind = 'gitlab';

UPDATE vcs_providers
SET api_url = 'https://next.forgejo.org'
WHERE vcs_kind = 'forgejo';

ALTER TABLE vcs_providers
ALTER COLUMN api_url SET NOT NULL;

---- create above / drop below ----
ALTER TABLE vcs_providers
	DROP COLUMN http_url,
	DROP COLUMN api_url;
