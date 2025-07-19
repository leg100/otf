ALTER TABLE vcs_providers
	ADD COLUMN http_url TEXT,
	ADD COLUMN api_url TEXT;
---- create above / drop below ----
ALTER TABLE vcs_providers
	DROP COLUMN http_url,
	DROP COLUMN api_url;
