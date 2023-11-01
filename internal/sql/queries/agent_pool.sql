-- name: InsertAgentPool :exec
INSERT INTO agent_pools (
    agent_pool_id,
    name,
    created_at,
    organization_name,
    organization_scoped
) VALUES (
    pggen.arg('agent_pool_id'),
    pggen.arg('name'),
    pggen.arg('created_at'),
    pggen.arg('organization_name'),
    pggen.arg('organization_scoped')
);

