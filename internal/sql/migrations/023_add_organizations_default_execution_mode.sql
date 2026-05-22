-- Add default execution kind column to organizations table, and populate with
-- default kind of 'remote' before setting not null.
--
-- Also add default agent pool id, which is optional and references the agent
-- pool id in the agent pools table.

ALTER TABLE organizations
    ADD COLUMN default_execution_kind TEXT,
    ADD COLUMN default_agent_pool_id TEXT;

UPDATE organizations
SET default_execution_kind = 'remote';

ALTER TABLE organizations
	ALTER COLUMN default_execution_kind SET NOT NULL;

ALTER TABLE organizations
	ADD CONSTRAINT organization_default_agent_pool_id_fkey FOREIGN KEY (default_agent_pool_id) REFERENCES agent_pools(agent_pool_id) ON UPDATE CASCADE;

-- Rename workspace column execution_mode to execution_kind to reflect new
-- naming in go codebase.

ALTER TABLE workspaces
	RENAME COLUMN execution_mode TO execution_kind;

---- create above / drop below ----

ALTER TABLE workspaces
	RENAME COLUMN execution_kind TO execution_mode;

ALTER TABLE organizations
    DROP COLUMN default_execution_kind,
    DROP COLUMN default_agent_pool_id;
