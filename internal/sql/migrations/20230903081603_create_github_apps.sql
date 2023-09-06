-- +goose Up
CREATE TABLE IF NOT EXISTS github_apps (
    github_app_id TEXT,
    app_id BIGINT,
    webhook_secret TEXT NOT NULL,
    private_key TEXT NOT NULL,
    organization_name TEXT REFERENCES organizations (name) ON UPDATE CASCADE ON DELETE CASCADE NOT NULL,
    PRIMARY KEY (github_app_id)
);

CREATE TABLE IF NOT EXISTS github_app_installs (
    github_app_install_id TEXT,
    install_id BIGINT,
    github_app_id TEXT REFERENCES github_apps (github_app_id) ON UPDATE CASCADE ON DELETE CASCADE NOT NULL,
    PRIMARY KEY (github_app_install_id),
    UNIQUE (github_app_install_id, github_app_id)
);

ALTER TABLE repo_connections
    ADD COLUMN github_app_install_id TEXT,
    ADD CONSTRAINT github_app_install_id_fk FOREIGN KEY (github_app_install_id)
        REFERENCES github_app_installs ON UPDATE CASCADE ON DELETE CASCADE,
    ALTER COLUMN webhook_id DROP NOT NULL,
    ADD CONSTRAINT webhook_check CHECK (github_app_install_id IS NOT NULL OR webhook_id IS NOT NULL);

ALTER TABLE vcs_providers
    ADD COLUMN github_app_install_id TEXT,
    ADD CONSTRAINT github_app_install_id_fk FOREIGN KEY (github_app_install_id)
        REFERENCES github_app_installs ON UPDATE CASCADE ON DELETE CASCADE,
    ALTER COLUMN token DROP NOT NULL,
    ADD CONSTRAINT auth_check CHECK (github_app_install_id IS NOT NULL OR token IS NOT NULL);

-- +goose Down
ALTER TABLE vcs_providers
    DROP COLUMN github_app_install_id,
    ALTER COLUMN token SET NOT NULL;

ALTER TABLE repo_connections
    DROP COLUMN github_app_install_id,
    ALTER COLUMN webhook_id SET NOT NULL;

DROP TABLE IF EXISTS github_app_installs;
DROP TABLE IF EXISTS github_apps;
