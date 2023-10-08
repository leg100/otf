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

-- add github app installs table, with fk to vcs providers; place mutually
-- exclusive constraint on user and org columns
CREATE TABLE IF NOT EXISTS github_app_installs (
    github_app_id BIGINT REFERENCES github_apps ON UPDATE CASCADE ON DELETE CASCADE NOT NULL,
    install_id BIGINT NOT NULL,
    username TEXT,
    organization TEXT,
    vcs_provider_id TEXT REFERENCES vcs_providers ON UPDATE CASCADE ON DELETE CASCADE NOT NULL,
    CHECK ((username IS NOT NULL AND organization IS NULL) OR (username IS NULL AND organization IS NOT NULL))
);

-- vcs provider token is no longer mandatory
ALTER TABLE vcs_providers
    ALTER COLUMN token DROP NOT NULL;

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

-- rename clouds to vcs_kinds and rename vcs_provider's fk
ALTER TABLE clouds RENAME TO vcs_kinds;
ALTER TABLE vcs_providers RENAME COLUMN cloud TO vcs_kind;

-- +goose Down

-- rename vcs_kinds back to clouds and rename fk back
ALTER TABLE vcs_kinds RENAME TO clouds;
ALTER TABLE vcs_providers RENAME COLUMN vcs_kind TO cloud;

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

-- make token mandatory on vcs providers
ALTER TABLE vcs_providers
    ALTER COLUMN token SET NOT NULL;

-- remove github tables
DROP TABLE IF EXISTS github_app_installs;
DROP TABLE IF EXISTS github_apps;
