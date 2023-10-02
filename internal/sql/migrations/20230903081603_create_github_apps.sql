-- +goose Up

-- add github apps table
CREATE TABLE IF NOT EXISTS github_apps (
    github_app_id BIGINT NOT NULL,
    webhook_secret TEXT NOT NULL,
    private_key TEXT NOT NULL,
    slug TEXT NOT NULL,
    organization TEXT,
    PRIMARY KEY (github_app_id)
);

-- add fk to vcs provider referencing github app id, along with github app
-- install id; token is not longer mandatory
ALTER TABLE vcs_providers
    ADD COLUMN github_app_id BIGINT,
    ADD CONSTRAINT github_app_id_fk FOREIGN KEY (github_app_id)
        REFERENCES github_apps ON UPDATE CASCADE ON DELETE CASCADE,
    ADD COLUMN github_app_install_id BIGINT,
    ALTER COLUMN token DROP NOT NULL,
    ADD CONSTRAINT auth_check CHECK (
        ((github_app_id IS NOT NULL AND github_app_install_id IS NOT NULL) AND token IS NULL)
        OR
        ((github_app_id IS NULL AND github_app_install_id IS NULL) AND token IS NOT NULL)
    );

-- alter repo_connections, swapping fk to webhooks with fk to vcs providers;
-- and copy repo path from webhooks to repo_connections
ALTER TABLE repo_connections
    ADD COLUMN repo_path TEXT,
    ADD COLUMN vcs_provider_id TEXT,
    ADD CONSTRAINT vcs_provider_id_fk FOREIGN KEY (vcs_provider_id)
        REFERENCES vcs_providers ON UPDATE CASCADE ON DELETE CASCADE;

UPDATE repo_connections rc
SET vcs_provider_id = w.vcs_provider_id,
    repo_path = w.identifier
FROM webhooks w
WHERE rc.webhook_id = w.webhook_id;

ALTER TABLE repo_connections
    ALTER COLUMN repo_path SET NOT NULL,
    ALTER COLUMN vcs_provider_id SET NOT NULL;

ALTER TABLE repo_connections
    DROP COLUMN webhook_id;

-- rename webhooks to repohooks, and rename columns
ALTER TABLE webhooks RENAME TO repohooks;
ALTER TABLE repohooks RENAME COLUMN webhook_id TO repohook_id;
ALTER TABLE repohooks RENAME COLUMN identifier TO repo_path;

-- +goose Down

-- rename repohooks back to webhooks, and rename columns back
ALTER TABLE repohooks RENAME TO webhooks;
ALTER TABLE webhooks RENAME COLUMN repohook_id TO webhook_id;
ALTER TABLE webhooks RENAME COLUMN repo_path TO identifier;

-- swap repo_connections fk from vcs providers back to webhooks
ALTER TABLE repo_connections
    ADD COLUMN webhook_id UUID;

UPDATE repo_connections rc
SET webhook_id = w.webhook_id
FROM webhooks w
WHERE rc.vcs_provider_id = w.vcs_provider_id;

ALTER TABLE repo_connections
    DROP COLUMN repo_path,
    DROP COLUMN vcs_provider_id,
    ALTER COLUMN webhook_id SET NOT NULL,
    ADD CONSTRAINT webhook_id_fk FOREIGN KEY (webhook_id)
        REFERENCES webhooks ON UPDATE CASCADE ON DELETE CASCADE;

-- remove github app id and with github app install id from vcs providers; make token mandatory
ALTER TABLE vcs_providers
    DROP COLUMN github_app_id,
    DROP COLUMN github_app_install_id,
    ALTER COLUMN token SET NOT NULL;

-- remove github tables
DROP TABLE IF EXISTS github_app_installs;
DROP TABLE IF EXISTS github_apps;
