-- +goose Up
-- +goose StatementBegin

-- create agent pools and agent_pool_allowed_workspaces tables
CREATE TABLE IF NOT EXISTS agent_pools (
    agent_pool_id       TEXT,
    name                TEXT NOT NULL,
    created_at          TIMESTAMPTZ NOT NULL,
    organization_name   TEXT REFERENCES organizations (name) ON UPDATE CASCADE ON DELETE CASCADE NOT NULL,
    organization_scoped BOOLEAN NOT NULL,
                        PRIMARY KEY (agent_pool_id)
);

CREATE TABLE IF NOT EXISTS agent_pool_allowed_workspaces (
    agent_pool_id TEXT REFERENCES agent_pools ON UPDATE CASCADE ON DELETE CASCADE NOT NULL,
    workspace_id  TEXT REFERENCES workspaces ON UPDATE CASCADE ON DELETE CASCADE NOT NULL,
    UNIQUE (agent_pool_id, workspace_id)
);

-- not necessary, but rename primary key to bring it into line with our standard approach of naming private keys according to the format <table_name>_id
ALTER TABLE agent_tokens
    RENAME COLUMN token_id TO agent_token_id;

-- alter agent tokens table, adding a fk to agent pools; for each organization
-- that has at least one agent token, add a default agent pool and update token
-- to reference that pool. Then drop the organization_name column.
ALTER TABLE agent_tokens
    ADD COLUMN agent_pool_id TEXT,
    ADD CONSTRAINT agent_pool_id_fk FOREIGN KEY (agent_pool_id)
        REFERENCES agent_pools ON UPDATE CASCADE ON DELETE CASCADE;

INSERT INTO agent_pools (agent_pool_id, name, created_at, organization_name, organization_scoped)
SELECT 'apool-' || substr(md5(random()::text), 0, 17), 'default', current_timestamp, o.name, true
FROM organizations o
JOIN agent_tokens at ON at.organization_name = o.name
GROUP BY o.name;

UPDATE agent_tokens at
SET agent_pool_id = ap.agent_pool_id
FROM agent_pools ap
WHERE ap.organization_name = at.organization_name;

ALTER TABLE agent_tokens
    ALTER COLUMN agent_pool_id SET NOT NULL,
    DROP COLUMN organization_name;

ALTER TABLE workspaces
    ADD COLUMN agent_pool_id TEXT,
    ADD CONSTRAINT agent_pool_fk FOREIGN KEY (agent_pool_id)
        REFERENCES agent_pools ON UPDATE CASCADE,
    ADD CONSTRAINT agent_pool_chk CHECK (execution_mode <> 'agent' OR agent_pool_id IS NOT NULL);

CREATE TABLE IF NOT EXISTS agent_statuses (
    status TEXT PRIMARY KEY
);

INSERT INTO agent_statuses (status) VALUES
	('busy'),
	('idle'),
	('exited'),
	('errored'),
	('unknown');

CREATE TABLE IF NOT EXISTS agents (
    agent_id       TEXT,
    name           TEXT,
    version        TEXT NOT NULL,
    concurrency    INT NOT NULL,
    ip_address     INET NOT NULL,
    last_ping_at   TIMESTAMPTZ NOT NULL,
    last_status_at TIMESTAMPTZ NOT NULL,
    status         TEXT REFERENCES agent_statuses ON UPDATE CASCADE NOT NULL,
    agent_pool_id  TEXT REFERENCES agent_pools ON UPDATE CASCADE ON DELETE CASCADE,
                   PRIMARY KEY (agent_id)
);

CREATE TABLE IF NOT EXISTS job_phases (
    phase TEXT PRIMARY KEY
);

INSERT INTO job_phases (phase) VALUES
	('plan'),
	('apply');

CREATE TABLE IF NOT EXISTS job_statuses (
    status TEXT PRIMARY KEY
);

INSERT INTO job_statuses (status) VALUES
	('unallocated'),
	('allocated'),
	('running'),
	('finished'),
	('errored'),
	('canceled');

CREATE TABLE IF NOT EXISTS signals (
    signal TEXT PRIMARY KEY
);

INSERT INTO signals (signal) VALUES
	('cancel'),
	('force_cancel');

CREATE TABLE IF NOT EXISTS jobs (
    run_id TEXT REFERENCES runs ON UPDATE CASCADE ON DELETE CASCADE NOT NULL,
    phase TEXT REFERENCES job_phases ON UPDATE CASCADE NOT NULL,
    status TEXT REFERENCES job_statuses ON UPDATE CASCADE NOT NULL,
    agent_id TEXT REFERENCES agents ON UPDATE CASCADE,
    signal TEXT REFERENCES signals ON UPDATE CASCADE
);

-- create triggers for pools, agents, and jobs
CREATE OR REPLACE FUNCTION agent_pools_notify_event() RETURNS TRIGGER AS $$
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
                      'id', record.agent_pool_id);
    PERFORM pg_notify('events', notification::text);
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION agents_notify_event() RETURNS TRIGGER AS $$
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
                      'id', record.agent_id);
    PERFORM pg_notify('events', notification::text);
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION jobs_notify_event() RETURNS TRIGGER AS $$
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
                      'id', record.run_id || '-' || record.phase);
    PERFORM pg_notify('events', notification::text);
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER notify_event
AFTER INSERT OR UPDATE OR DELETE ON agent_pools
    FOR EACH ROW EXECUTE PROCEDURE agent_pools_notify_event();

CREATE TRIGGER notify_event
AFTER INSERT OR UPDATE OR DELETE ON agents
    FOR EACH ROW EXECUTE PROCEDURE agents_notify_event();

CREATE TRIGGER notify_event
AFTER INSERT OR UPDATE OR DELETE ON jobs
    FOR EACH ROW EXECUTE PROCEDURE jobs_notify_event();
-- +goose StatementEnd

-- +goose Down
DROP TRIGGER IF EXISTS notify_event ON jobs;
DROP TRIGGER IF EXISTS notify_event ON agents;
DROP TRIGGER IF EXISTS notify_event ON agent_pools;
DROP FUNCTION IF EXISTS jobs_notify_event;
DROP FUNCTION IF EXISTS agents_notify_event;
DROP FUNCTION IF EXISTS agent_pools_notify_event;

DROP TABLE IF EXISTS jobs;
DROP TABLE IF EXISTS job_statuses;
DROP TABLE IF EXISTS signals;
DROP TABLE IF EXISTS job_phases;

DROP TABLE IF EXISTS agents;
DROP TABLE IF EXISTS agent_statuses;

ALTER TABLE workspaces
    DROP COLUMN agent_pool_id;

-- for each agent token, lookup its organization via its agent pool and set
-- that as its organization. Then making the organization_name column not
-- null and drop the agent_pool_id column.
ALTER TABLE agent_tokens
    ADD COLUMN organization_name TEXT;

UPDATE agent_tokens at
SET organization_name = ap.organization_name
FROM agent_pools ap
WHERE at.agent_pool_id = ap.agent_pool_id;

ALTER TABLE agent_tokens
    ALTER COLUMN organization_name SET NOT NULL,
    DROP COLUMN agent_pool_id;

ALTER TABLE agent_tokens
    RENAME COLUMN agent_token_id TO token_id;

DROP TABLE IF EXISTS agent_pool_allowed_workspaces;
DROP TABLE IF EXISTS agent_pools;
