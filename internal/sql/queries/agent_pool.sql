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

-- name: FindAgentPools :many
SELECT ap.*,
    (
        SELECT array_agg(w.workspace_id)
        FROM workspaces w
        WHERE w.agent_pool_id = ap.agent_pool_id
    ) AS workspace_ids,
    (
        SELECT array_agg(aw.workspace_id)
        FROM agent_pool_allowed_workspaces aw
        WHERE aw.agent_pool_id = ap.agent_pool_id
    ) AS allowed_workspace_ids
FROM agent_pools ap
LEFT JOIN (agent_pool_allowed_workspaces aw JOIN workspaces w USING (workspace_id)) ON ap.agent_pool_id = aw.agent_pool_id
WHERE ((pggen.arg('organization_name')::text IS NULL) OR ap.organization_name = pggen.arg('organization_name'))
AND   ((pggen.arg('name_substring')::text IS NULL) OR ap.name LIKE '%' || pggen.arg('name_substring') || '%')
AND   ((pggen.arg('allowed_workspace_name')::text IS NULL) OR
        ap.organization_scoped OR w.name = pggen.arg('allowed_workspace_name')
      )
GROUP BY ap.agent_pool_id
;

-- name: FindAgentPool :one
SELECT ap.*,
    (
        SELECT array_agg(w.workspace_id)
        FROM workspaces w
        WHERE w.agent_pool_id = ap.agent_pool_id
    ) AS workspace_ids,
    (
        SELECT array_agg(aw.workspace_id)
        FROM agent_pool_allowed_workspaces aw
        WHERE aw.agent_pool_id = ap.agent_pool_id
    ) AS allowed_workspace_ids
FROM agent_pools ap
WHERE ap.agent_pool_id = pggen.arg('pool_id')
GROUP BY ap.agent_pool_id
;

-- name: FindAgentPoolByAgentTokenID :one
SELECT ap.*,
    (
        SELECT array_agg(w.workspace_id)
        FROM workspaces w
        WHERE w.agent_pool_id = ap.agent_pool_id
    ) AS workspace_ids,
    (
        SELECT array_agg(aw.workspace_id)
        FROM agent_pool_allowed_workspaces aw
        WHERE aw.agent_pool_id = ap.agent_pool_id
    ) AS allowed_workspace_ids
FROM agent_pools ap
JOIN agent_tokens at USING (agent_pool_id)
WHERE at.agent_token_id = pggen.arg('agent_token_id')
GROUP BY ap.agent_pool_id
;

-- name: UpdateAgentPool :one
UPDATE agent_pools
SET name = pggen.arg('name'),
    organization_scoped = pggen.arg('organization_scoped')
WHERE agent_pool_id = pggen.arg('pool_id')
RETURNING *;

-- name: DeleteAgentPool :one
DELETE
FROM agent_pools
WHERE agent_pool_id = pggen.arg('pool_id')
RETURNING organization_name
;

-- name: InsertAgentPoolAllowedWorkspaces :exec
INSERT INTO agent_pool_allowed_workspaces (
    agent_pool_id,
    workspace_id
) VALUES (
    pggen.arg('pool_id'),
    pggen.arg('workspace_id')
);

-- name: DeleteAgentPoolAllowedWorkspaces :exec
DELETE
FROM agent_pool_allowed_workspaces
WHERE agent_pool_id = pggen.arg('pool_id')
;
