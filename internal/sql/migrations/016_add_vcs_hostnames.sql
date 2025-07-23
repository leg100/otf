-- Add three fields to vcs_providers:

-- * api_url
-- * base_url
-- * tfe_service_provider

-- They're all mandatory, so add, then provided default values, then set
-- columns not null

ALTER TABLE vcs_providers
	ADD COLUMN base_url TEXT,
	ADD COLUMN api_url TEXT,
	ADD COLUMN tfe_service_provider TEXT;

UPDATE vcs_providers
SET
	base_url = 'https://github.com',
	api_url = 'https://github.com',
	tfe_service_provider = 'github'
WHERE vcs_kind = 'github-token';

UPDATE vcs_providers
SET
	base_url = 'https://github.com',
	api_url = 'https://github.com',
	tfe_service_provider = 'github_app'
WHERE vcs_kind = 'github-app';

UPDATE vcs_providers
SET
	base_url = 'https://gitlab.com',
	api_url = 'https://gitlab.com',
	tfe_service_provider = 'gitlab_hosted'
WHERE vcs_kind = 'gitlab';

UPDATE vcs_providers
SET
	base_url = 'https://next.forgejo.org',
	api_url = 'https://next.forgejo.org',
	tfe_service_provider = 'forgejo'
WHERE vcs_kind = 'forgejo';

ALTER TABLE vcs_providers
	ALTER COLUMN base_url SET NOT NULL,
	ALTER COLUMN api_url SET NOT NULL,
	ALTER COLUMN tfe_service_provider SET NOT NULL;

---- create above / drop below ----
ALTER TABLE vcs_providers
	DROP COLUMN base_url,
	DROP COLUMN api_url,
	DROP COLUMN tfe_service_provider;
