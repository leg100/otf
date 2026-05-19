-- Add default execution mode column to organizations table, and populate with
-- default mode of 'remote' before setting not null.
--
-- Also add default agent pool id, which is optional and references the agent
-- pool id in the agent pools table.

ALTER TABLE organizations
    ADD COLUMN default_execution_mode TEXT,
    ADD COLUMN default_agent_pool_id TEXT;

UPDATE organizations
SET default_execution_mode = 'remote';

ALTER TABLE organizations
	ALTER COLUMN default_execution_mode SET NOT NULL;

ALTER TABLE organizations
	ADD CONSTRAINT organization_default_agent_pool_id_fkey FOREIGN KEY (default_agent_pool_id) REFERENCES agent_pools(agent_pool_id) ON UPDATE CASCADE;

---- create above / drop below ----

ALTER TABLE organizations
    DROP COLUMN default_execution_mode,
    DROP COLUMN default_agent_pool_id;
