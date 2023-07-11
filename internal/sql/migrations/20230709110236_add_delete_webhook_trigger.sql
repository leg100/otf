-- +goose Up
-- +goose StatementBegin

-- add ON DELETE CASCADE to repo_connections, to ensure they are deleted
-- whenever anything they reference is deleted
ALTER TABLE repo_connections
    DROP CONSTRAINT repo_connections_module_id_fkey,
    DROP CONSTRAINT repo_connections_workspace_id_fkey,
    DROP CONSTRAINT repo_connections_webhook_id_fkey,
    ADD CONSTRAINT repo_connections_module_id_fkey FOREIGN KEY (module_id)
        REFERENCES modules ON UPDATE CASCADE ON DELETE CASCADE,
    ADD CONSTRAINT repo_connections_workspace_id_fkey FOREIGN KEY (workspace_id)
        REFERENCES workspaces ON UPDATE CASCADE ON DELETE CASCADE,
    ADD CONSTRAINT repo_connections_webhook_id_fkey FOREIGN KEY (webhook_id)
        REFERENCES webhooks ON UPDATE CASCADE ON DELETE CASCADE;

--
-- move vcs_provider_id column from repo_connections to webhooks and add ON
-- DELETE CASCADE, to ensure webhooks are deleted whenever their vcs_providers
-- are deleted
--
ALTER TABLE webhooks
    ADD COLUMN vcs_provider_id TEXT,
    ADD CONSTRAINT vcs_provider_id_fk FOREIGN KEY (vcs_provider_id)
        REFERENCES vcs_providers ON UPDATE CASCADE ON DELETE CASCADE,
    DROP COLUMN cloud,
    ADD CONSTRAINT webhooks_cloud_id_uniq
        UNIQUE(identifier, vcs_provider_id);

UPDATE webhooks w
SET vcs_provider_id = c.vcs_provider_id
FROM repo_connections c
WHERE w.webhook_id = c.webhook_id;

ALTER TABLE webhooks ALTER COLUMN vcs_provider_id SET NOT NULL;
ALTER TABLE repo_connections DROP COLUMN vcs_provider_id;

CREATE OR REPLACE FUNCTION repo_connections_notify_event() RETURNS TRIGGER AS $$
DECLARE
    record RECORD;
    notification JSON;
BEGIN
    IF (TG_OP = 'DELETE') THEN
        record = OLD;
    ELSE
        record = NEW;
    END IF;
    notification = json_build_object(
                      'table',TG_TABLE_NAME,
                      'action', TG_OP,
                      'id', record.webhook_id);
    PERFORM pg_notify('events', notification::text);
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER notify_event
AFTER DELETE ON repo_connections
    FOR EACH ROW EXECUTE PROCEDURE repo_connections_notify_event();

-- +goose StatementEnd

-- +goose Down
ALTER TABLE repo_connections
    DROP CONSTRAINT repo_connections_module_id_fkey,
    DROP CONSTRAINT repo_connections_workspace_id_fkey,
    DROP CONSTRAINT repo_connections_webhook_id_fkey,
    ADD CONSTRAINT repo_connections_module_id_fkey FOREIGN KEY (module_id)
        REFERENCES modules ON UPDATE CASCADE,
    ADD CONSTRAINT repo_connections_workspace_id_fkey FOREIGN KEY (workspace_id)
        REFERENCES workspaces ON UPDATE CASCADE,
    ADD CONSTRAINT repo_connections_webhook_id_fkey FOREIGN KEY (webhook_id)
        REFERENCES webhooks ON UPDATE CASCADE;

ALTER TABLE repo_connections
    ADD COLUMN vcs_provider_id TEXT NOT NULL,
    ADD CONSTRAINT vcs_provider_id_fk FOREIGN KEY (vcs_provider_id) REFERENCES vcs_providers ON UPDATE CASCADE;

UPDATE repo_connections c
SET vcs_provider_id = w.vcs_provider_id
FROM webhooks w
WHERE c.webhook_id = w.webhook_id;

ALTER TABLE webhooks
    DROP COLUMN vcs_provider_id,
    ADD COLUMN cloud TEXT NOT NULL,
    ADD CONSTRAINT webhooks_cloud_fkey FOREIGN KEY (cloud)
        REFERENCES clouds ON UPDATE CASCADE ON DELETE CASCADE,
    ADD CONSTRAINT webhooks_cloud_id_uniq UNIQUE(cloud, identifier);
