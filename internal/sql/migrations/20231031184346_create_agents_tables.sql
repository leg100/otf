-- +goose Up
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
    workspace_id  TEXT REFERENCES workspaces ON UPDATE CASCADE ON DELETE CASCADE NOT NULL
);

-- alter agent tokens table, adding a fk to agent pools; for each organization
-- that has at least one agent token, add a default agent pool and update token
-- to reference that pool.
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
    ALTER COLUMN agent_pool_id SET NOT NULL;

ALTER TABLE workspaces
    ADD COLUMN agent_pool_id TEXT,
    ADD CONSTRAINT agent_pool_id_fk FOREIGN KEY (agent_pool_id)
        REFERENCES agent_pools ON UPDATE CASCADE;

CREATE TABLE IF NOT EXISTS agent_statuses (
    status TEXT PRIMARY KEY
);

INSERT INTO agent_statuses (status) VALUES
	('busy'),
	('idle'),
	('exited'),
	('errored');

CREATE TABLE IF NOT EXISTS agents (
    agent_id       TEXT,
    name           TEXT,
    concurrency    INT NOT NULL,
    server         BOOLEAN NOT NULL,
    ip_address     TEXT NOT NULL,
    last_ping_at   TIMESTAMPTZ NOT NULL,
    status         TEXT REFERENCES agent_statuses ON UPDATE CASCADE NOT NULL,
    agent_token_id TEXT REFERENCES agent_tokens ON UPDATE CASCADE ON DELETE CASCADE,
                   PRIMARY KEY (agent_id)
                   CHECK (server OR agent_token_id IS NOT NULL)
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
	('pending'),
	('running'),
	('errored');

CREATE TABLE IF NOT EXISTS jobs (
    run_id TEXT REFERENCES runs ON UPDATE CASCADE ON DELETE CASCADE NOT NULL,
    phase TEXT REFERENCES job_phases ON UPDATE CASCADE NOT NULL,
    status TEXT REFERENCES job_statuses ON UPDATE CASCADE NOT NULL,
    agent_id TEXT REFERENCES agents ON UPDATE CASCADE ON DELETE CASCADE
);

-- +goose Down
DROP TABLE IF EXISTS jobs;
DROP TABLE IF EXISTS job_statuses;
DROP TABLE IF EXISTS job_phases;

DROP TABLE IF EXISTS agents;
DROP TABLE IF EXISTS agent_statuses;

ALTER TABLE workspaces
    DROP COLUMN agent_pool_id;

ALTER TABLE agent_tokens
    DROP COLUMN agent_pool_id;

DROP TABLE IF EXISTS agent_pool_allowed_workspaces;
DROP TABLE IF EXISTS agent_pools;
