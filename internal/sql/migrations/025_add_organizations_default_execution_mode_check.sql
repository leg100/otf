-- Enforce at the database level that an organization specifying an agent
-- default execution kind also specifies a default agent pool.
--
-- First repair any rows left inconsistent, if necessary.

UPDATE organizations
SET default_execution_kind = 'remote',
    default_agent_pool_id = NULL
WHERE default_execution_kind NOT IN ('remote', 'local', 'agent')
OR    (default_execution_kind = 'agent' AND default_agent_pool_id IS NULL)
OR    (default_execution_kind <> 'agent' AND default_agent_pool_id IS NOT NULL);

ALTER TABLE organizations
    ADD CONSTRAINT organizations_default_agent_pool_chk
    CHECK ((default_execution_kind = 'agent') = (default_agent_pool_id IS NOT NULL));

---- create above / drop below ----

ALTER TABLE organizations
    DROP CONSTRAINT organizations_default_agent_pool_chk;
