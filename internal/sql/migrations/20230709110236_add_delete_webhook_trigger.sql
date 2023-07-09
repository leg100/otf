-- +goose Up
-- +goose StatementBegin
ALTER TABLE webhooks
    ADD COLUMN vcs_provider_id TEXT NOT NULL,
    ADD CONSTRAINT vcs_provider_id_fk FOREIGN KEY (vcs_provider_id) REFERENCES vcs_providers ON UPDATE CASCADE;

UPDATE webhooks w
SET vcs_provider_id = c.vcs_provider_id
FROM repo_connections c
WHERE w.webhook_id = c.webhook_id;

ALTER TABLE repo_connections DROP COLUMN vcs_provider_id;

CREATE OR REPLACE FUNCTION delete_webhooks() RETURNS TRIGGER AS $$
BEGIN
    DELETE
    FROM webhooks w
    WHERE NOT EXISTS (
        SELECT FROM repo_connections c
        WHERE c.webhook_id = w.webhook_id
    );
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- invoke trigger to remove existing unreferenced webhooks
DELETE FROM repo_connections WHERE workspace_id = '--invalid-name--';

CREATE TRIGGER delete_webhooks
AFTER DELETE ON repo_connections
    FOR EACH STATEMENT EXECUTE FUNCTION delete_webhooks();

CREATE OR REPLACE FUNCTION webhook_notify_event() RETURNS TRIGGER AS $$
DECLARE
    record RECORD;
    notification JSON;
BEGIN
    record = OLD;
    notification = json_build_object(
                      'table',TG_TABLE_NAME,
                      'action', TG_OP,
                      'id', record.webhook_id,
                      'payload', to_json(record));
    PERFORM pg_notify('events', notification::text);
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER notify_event
AFTER DELETE ON webhooks
    FOR EACH ROW EXECUTE PROCEDURE webhook_notify_event();
-- +goose StatementEnd

-- +goose Down
DROP TRIGGER IF EXISTS delete_webhooks ON repo_connections;
DROP FUNCTION IF EXISTS delete_webhooks;
DROP TRIGGER IF EXISTS notify_event ON webhooks;
DROP FUNCTION IF EXISTS webhook_notify_event;

ALTER TABLE repo_connections
    ADD COLUMN vcs_provider_id TEXT NOT NULL,
    ADD CONSTRAINT vcs_provider_id_fk FOREIGN KEY (vcs_provider_id) REFERENCES vcs_providers ON UPDATE CASCADE;

UPDATE repo_connections c
SET vcs_provider_id = w.vcs_provider_id
FROM webhooks w
WHERE c.webhook_id = w.webhook_id;

ALTER TABLE webhooks DROP COLUMN vcs_provider_id;
