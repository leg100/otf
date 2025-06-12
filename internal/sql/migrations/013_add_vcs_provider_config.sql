-- Refactor VCS provider tables.

-- Add a new VCS kind 'github-app' to differentiate it from github kinds that
-- use a token.
INSERT INTO vcs_kinds VALUES ('github-app');

-- Merge github_app_installs into vcs_providers.
ALTER TABLE vcs_providers RENAME COLUMN github_app_id TO install_app_id;
ALTER TABLE vcs_providers
	ADD COLUMN install_id BIGINT,
	ADD COLUMN install_username TEXT,
	ADD COLUMN install_organization TEXT;

UPDATE vcs_providers vcs
SET install_id = gai.install_id,
	install_username = gai.username,
	install_organization = gai.organization,
	vcs_kind = 'github-app'
FROM github_app_installs gai
WHERE gai.vcs_provider_id = vcs.vcs_provider_id;

DROP TABLE github_app_installs;

-- Rename github kind to differentiate it from github app kind.
UPDATE vcs_kinds
SET name = 'github-token'
WHERE name = 'github';

---- create above / drop below ----

-- Rename back the vcs kinds github-* to github
UPDATE vcs_kinds
SET name = 'github'
WHERE name = 'github-token';

UPDATE vcs_providers
SET vcs_kind = 'github'
WHERE vcs_kind = 'github-app';

DELETE FROM vcs_kinds
WHERE name = 'github-app';

-- Add back the github app installs table and re-populate
CREATE TABLE IF NOT EXISTS github_app_installs (
    github_app_id bigint NOT NULL,
    install_id bigint NOT NULL,
    username text,
    organization text,
    vcs_provider_id text NOT NULL,
    CONSTRAINT github_app_installs_check CHECK ((((username IS NOT NULL) AND (organization IS NULL)) OR ((username IS NULL) AND (organization IS NOT NULL)))),
    CONSTRAINT github_app_installs_github_app_id_fkey FOREIGN KEY (github_app_id) REFERENCES github_apps(github_app_id) ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT github_app_installs_vcs_provider_id_fkey FOREIGN KEY (vcs_provider_id) REFERENCES vcs_providers(vcs_provider_id) ON UPDATE CASCADE ON DELETE CASCADE
);
INSERT INTO github_app_installs (
	github_app_id,
	install_id,
	username,
	organization,
	vcs_provider_id
)
SELECT
	install_app_id,
	install_id,
	install_username,
	install_organization,
	vcs_provider_id
FROM vcs_providers
WHERE install_id IS NOT NULL;

-- Revert the vcs_providers table back to how it was before.
ALTER TABLE vcs_providers RENAME COLUMN install_app_id TO github_app_id;
ALTER TABLE vcs_providers
	DROP COLUMN install_id,
	DROP COLUMN install_username,
	DROP COLUMN install_organization;
