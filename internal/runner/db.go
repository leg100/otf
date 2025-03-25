package runner

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/sql"
)

type db struct {
	*sql.DB
}

func (db *db) create(ctx context.Context, meta *RunnerMeta) error {
	args := pgx.NamedArgs{
		"runner_id":      meta.ID,
		"name":           meta.Name,
		"version":        meta.Version,
		"max_jobs":       meta.MaxJobs,
		"ip_address":     meta.IPAddress,
		"last_ping_at":   meta.LastPingAt,
		"last_status_at": meta.LastStatusAt,
		"status":         meta.Status,
	}
	if meta.AgentPool != nil {
		args["agent_pool_id"] = meta.AgentPool.ID
	}
	_, err := db.Exec(ctx, `
INSERT INTO runners (
    runner_id,
    name,
    version,
    max_jobs,
    ip_address,
    last_ping_at,
    last_status_at,
    status,
    agent_pool_id
) VALUES (
	@runner_id,
	@name,
	@version,
	@max_jobs,
	@ip_address,
	@last_ping_at,
	@last_status_at,
	@status,
	@agent_pool_id
)`, args)
	return err
}

func (db *db) update(ctx context.Context, runnerID resource.TfeID, fn func(context.Context, *RunnerMeta) error) error {
	_, err := sql.Updater(
		ctx,
		db.DB,
		func(ctx context.Context, conn sql.Connection) (*RunnerMeta, error) {
			rows := db.Query(ctx, `
SELECT
    a.runner_id, a.name, a.version, a.max_jobs, a.ip_address, a.last_ping_at, a.last_status_at, a.status,
    ap::"agent_pools" AS agent_pool,
    ( SELECT count(*)
      FROM jobs j
      WHERE a.runner_id = j.runner_id
      AND j.status IN ('allocated', 'running')
    ) AS current_jobs
FROM runners a
LEFT JOIN agent_pools ap USING (agent_pool_id)
WHERE a.runner_id = $1
FOR UPDATE OF a
`, runnerID)
			return sql.CollectOneRow(rows, scanRunner)
		},
		fn,
		func(ctx context.Context, conn sql.Connection, agent *RunnerMeta) error {
			_, err := db.Exec(ctx, `
UPDATE runners
SET status = @status,
    last_ping_at = @last_ping_at,
    last_status_at = @last_status_at
WHERE runner_id = @runner_id
`,
				pgx.NamedArgs{
					"status":         agent.Status,
					"last_ping_at":   agent.LastPingAt,
					"last_status_at": agent.LastStatusAt,
					"runner_id":      agent.ID,
				})
			return err
		},
	)
	return err
}

func (db *db) get(ctx context.Context, runnerID resource.TfeID) (*RunnerMeta, error) {
	rows := db.Query(ctx, `
SELECT
    a.runner_id, a.name, a.version, a.max_jobs, a.ip_address, a.last_ping_at, a.last_status_at, a.status,
    ap::"agent_pools" AS agent_pool,
    ( SELECT count(*)
      FROM jobs j
      WHERE a.runner_id = j.runner_id
      AND j.status IN ('allocated', 'running')
    ) AS current_jobs
FROM runners a
LEFT JOIN agent_pools ap USING (agent_pool_id)
WHERE a.runner_id = $1
`, runnerID)
	return sql.CollectOneRow(rows, scanRunner)
}

func (db *db) list(ctx context.Context) ([]*RunnerMeta, error) {
	rows := db.Query(ctx, `
SELECT
    a.runner_id, a.name, a.version, a.max_jobs, a.ip_address, a.last_ping_at, a.last_status_at, a.status,
    ap::"agent_pools" AS agent_pool,
    ( SELECT count(*)
      FROM jobs j
      WHERE a.runner_id = j.runner_id
      AND j.status IN ('allocated', 'running')
    ) AS current_jobs
FROM runners a
LEFT JOIN agent_pools ap USING (agent_pool_id)
ORDER BY a.last_ping_at DESC
`)
	return sql.CollectRows(rows, scanRunner)
}

func (db *db) listServerRunners(ctx context.Context) ([]*RunnerMeta, error) {
	rows := db.Query(ctx, `
SELECT
    a.runner_id, a.name, a.version, a.max_jobs, a.ip_address, a.last_ping_at, a.last_status_at, a.status,
    ap::"agent_pools" AS agent_pool,
    ( SELECT count(*)
      FROM jobs j
      WHERE a.runner_id = j.runner_id
      AND j.status IN ('allocated', 'running')
    ) AS current_jobs
FROM runners a
LEFT JOIN agent_pools ap USING (agent_pool_id)
WHERE agent_pool_id IS NULL
ORDER BY last_ping_at DESC
`)
	return sql.CollectRows(rows, pgx.RowToFunc[*RunnerMeta](scanRunner))
}

func (db *db) listRunnersByOrganization(ctx context.Context, organization organization.Name) ([]*RunnerMeta, error) {
	rows := db.Query(ctx, `
SELECT
    a.runner_id, a.name, a.version, a.max_jobs, a.ip_address, a.last_ping_at, a.last_status_at, a.status,
    ap::"agent_pools" AS agent_pool,
    ( SELECT count(*)
      FROM jobs j
      WHERE a.runner_id = j.runner_id
      AND j.status IN ('allocated', 'running')
    ) AS current_jobs
FROM runners a
JOIN agent_pools ap USING (agent_pool_id)
WHERE ap.organization_name = $1
ORDER BY last_ping_at DESC
`, organization)
	return sql.CollectRows(rows, scanRunner)
}

func (db *db) listRunnersByPool(ctx context.Context, poolID resource.TfeID) ([]*RunnerMeta, error) {
	rows := db.Query(ctx, `
SELECT
    a.runner_id, a.name, a.version, a.max_jobs, a.ip_address, a.last_ping_at, a.last_status_at, a.status,
    ap::"agent_pools" AS agent_pool,
    ( SELECT count(*)
      FROM jobs j
      WHERE a.runner_id = j.runner_id
      AND j.status IN ('allocated', 'running')
    ) AS current_jobs
FROM runners a
JOIN agent_pools ap USING (agent_pool_id)
WHERE ap.agent_pool_id = $1
ORDER BY last_ping_at DESC
`, poolID)
	return sql.CollectRows(rows, scanRunner)
}

func (db *db) deleteRunner(ctx context.Context, runnerID resource.TfeID) error {
	_, err := db.Exec(ctx, `
DELETE
FROM runners
WHERE runner_id = $1
RETURNING runner_id, name, version, max_jobs, ip_address, last_ping_at, last_status_at, status, agent_pool_id
`, runnerID)
	return err
}

func scanRunner(row pgx.CollectableRow) (*RunnerMeta, error) {
	runner, err := pgx.RowToAddrOfStructByName[RunnerMeta](row)
	if err != nil {
		return nil, err
	}
	runner.LastPingAt = runner.LastPingAt.UTC()
	runner.LastStatusAt = runner.LastStatusAt.UTC()
	return runner, nil
}

// jobs

func (db *db) createJob(ctx context.Context, job *Job) error {
	_, err := db.Exec(ctx, `
INSERT INTO jobs (
    job_id,
    run_id,
    phase,
    status
) VALUES (
    $1,
    $2,
    $3,
    $4
)`,
		job.ID,
		job.RunID,
		job.Phase,
		job.Status,
	)
	return err
}

func (db *db) getAllocatedAndSignaledJobs(ctx context.Context, runnerID resource.TfeID) ([]*Job, error) {
	rows := db.Query(ctx, `
SELECT
    j.job_id,
    j.run_id,
    j.phase,
    j.status,
    j.signaled,
    j.runner_id,
    w.agent_pool_id,
    r.workspace_id,
    w.organization_name
FROM jobs j
JOIN runs r USING (run_id)
JOIN workspaces w USING (workspace_id)
WHERE j.runner_id = $1
AND   j.status = 'allocated'
`, runnerID)
	allocated, err := sql.CollectRows(rows, pgx.RowToAddrOfStructByName[Job])
	if err != nil {
		return nil, err
	}
	rows = db.Query(ctx, `
UPDATE jobs AS j
SET signaled = NULL
FROM runs r, workspaces w
WHERE j.run_id = r.run_id
AND   r.workspace_id = w.workspace_id
AND   j.runner_id = $1
AND   j.status = 'running'
AND   j.signaled IS NOT NULL
RETURNING
    j.job_id,
    j.run_id,
    j.phase,
    j.status,
    j.signaled,
    j.runner_id,
    w.agent_pool_id,
    r.workspace_id,
    w.organization_name
`, runnerID)
	signaled, err := sql.CollectRows(rows, pgx.RowToAddrOfStructByName[Job])
	if err != nil {
		return nil, err
	}
	return append(allocated, signaled...), nil
}

func (db *db) getJob(ctx context.Context, jobID resource.TfeID) (*Job, error) {
	rows := db.Query(ctx, `
SELECT
    j.job_id,
    j.run_id,
    j.phase,
    j.status,
    j.signaled,
    j.runner_id,
    w.agent_pool_id,
    r.workspace_id,
    w.organization_name
FROM jobs j
JOIN runs r USING (run_id)
JOIN workspaces w USING (workspace_id)
WHERE j.job_id = $1
`, jobID)
	return sql.CollectOneRow(rows, pgx.RowToAddrOfStructByName[Job])
}

func (db *db) listJobs(ctx context.Context) ([]*Job, error) {
	rows := db.Query(ctx, `
SELECT
    j.job_id,
    j.run_id,
    j.phase,
    j.status,
    j.signaled,
    j.runner_id,
    w.agent_pool_id,
    r.workspace_id,
    w.organization_name
FROM jobs j
JOIN runs r USING (run_id)
JOIN workspaces w USING (workspace_id)
`)
	return sql.CollectRows(rows, pgx.RowToAddrOfStructByName[Job])
}

func (db *db) updateJob(ctx context.Context, jobID resource.TfeID, fn func(context.Context, *Job) error) (*Job, error) {
	return sql.Updater(
		ctx,
		db.DB,
		func(ctx context.Context, conn sql.Connection) (*Job, error) {
			rows := db.Query(ctx, `
SELECT
    j.job_id,
    j.run_id,
    j.phase,
    j.status,
    j.signaled,
    j.runner_id,
    w.agent_pool_id,
    r.workspace_id,
    w.organization_name
FROM jobs j
JOIN runs r USING (run_id)
JOIN workspaces w USING (workspace_id)
WHERE j.job_id = $1
FOR UPDATE OF j
`, jobID)
			return sql.CollectOneRow(rows, pgx.RowToAddrOfStructByName[Job])
		},
		fn,
		func(ctx context.Context, conn sql.Connection, job *Job) error {
			_, err := db.Exec(ctx, `
UPDATE jobs
SET status   = $1,
    signaled = $2,
    runner_id = $3
WHERE job_id = $4
RETURNING run_id, phase, status, runner_id, signaled, job_id
`,
				job.Status,
				job.Signaled,
				job.RunnerID,
				job.ID,
			)
			return err
		},
	)
}

func (db *db) updateUnfinishedJobByRunID(ctx context.Context, runID resource.TfeID, fn func(context.Context, *Job) error) (*Job, error) {
	return sql.Updater(
		ctx,
		db.DB,
		func(ctx context.Context, conn sql.Connection) (*Job, error) {
			rows := db.Query(ctx, `
SELECT
    j.job_id,
    j.run_id,
    j.phase,
    j.status,
    j.signaled,
    j.runner_id,
    w.agent_pool_id,
    r.workspace_id,
    w.organization_name
FROM jobs j
JOIN runs r USING (run_id)
JOIN workspaces w USING (workspace_id)
WHERE j.run_id = $1
AND   j.status IN ('unallocated', 'allocated', 'running')
FOR UPDATE OF j
`, runID)
			return sql.CollectOneRow(rows, pgx.RowToAddrOfStructByName[Job])
		},
		fn,
		func(ctx context.Context, conn sql.Connection, job *Job) error {
			_, err := db.Exec(ctx, `
UPDATE jobs
SET status   = $1,
    signaled = $2,
    runner_id = $3
WHERE job_id = $4
`,
				job.Status,
				job.Signaled,
				job.RunnerID,
				job.ID,
			)
			return err
		},
	)
}

// agent tokens

func (db *db) createAgentToken(ctx context.Context, token *agentToken) error {
	_, err := db.Exec(ctx, `
INSERT INTO agent_tokens (
    agent_token_id,
    created_at,
    description,
    agent_pool_id
) VALUES (
    $1,
    $2,
    $3,
    $4
)`,
		token.ID,
		token.CreatedAt,
		token.Description,
		token.AgentPoolID,
	)
	return err
}

func (db *db) getAgentTokenByID(ctx context.Context, id resource.TfeID) (*agentToken, error) {
	rows := db.Query(ctx, `
SELECT agent_token_id, created_at, description, agent_pool_id
FROM agent_tokens
WHERE agent_token_id = $1
`, id)
	return sql.CollectOneRow(rows, scanAgentToken)
}

func (db *db) listAgentTokens(ctx context.Context, poolID resource.TfeID) ([]*agentToken, error) {
	rows := db.Query(ctx, `
SELECT agent_token_id, created_at, description, agent_pool_id
FROM agent_tokens
WHERE agent_pool_id = $1
ORDER BY created_at DESC
`, poolID)
	return sql.CollectRows(rows, scanAgentToken)
}

func (db *db) deleteAgentToken(ctx context.Context, id resource.TfeID) error {
	_, err := db.Exec(ctx, `
DELETE
FROM agent_tokens
WHERE agent_token_id = $1
`, id)
	return err
}

func scanAgentToken(row pgx.CollectableRow) (*agentToken, error) {
	token, err := pgx.RowToAddrOfStructByName[agentToken](row)
	if err != nil {
		return nil, err
	}
	token.CreatedAt = token.CreatedAt.UTC()
	return token, nil
}

// agent pools

func (db *db) createPool(ctx context.Context, pool *Pool) error {
	err := db.Tx(ctx, func(ctx context.Context, conn sql.Connection) error {
		_, err := db.Exec(ctx, `
INSERT INTO agent_pools (
    agent_pool_id,
    name,
    created_at,
    organization_name,
    organization_scoped
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5
)`,
			pool.ID,
			pool.Name,
			pool.CreatedAt,
			pool.Organization,
			pool.OrganizationScoped,
		)
		if err != nil {
			return err
		}
		for _, workspaceID := range pool.AllowedWorkspaces {
			_, err := db.Exec(ctx, `
INSERT INTO agent_pool_allowed_workspaces (
    agent_pool_id,
    workspace_id
) VALUES (
    $1,
    $2
)`, pool.ID, workspaceID)
			if err != nil {
				return err
			}
		}
		return nil
	})
	return err
}

func (db *db) updatePool(ctx context.Context, pool *Pool) error {
	_, err := db.Exec(ctx, `
UPDATE agent_pools
SET name = $1,
    organization_scoped = $2
WHERE agent_pool_id = $3
RETURNING agent_pool_id, name, created_at, organization_name, organization_scoped
`,
		pool.Name,
		pool.OrganizationScoped,
		pool.ID,
	)
	return err
}

func (db *db) addAgentPoolAllowedWorkspace(ctx context.Context, poolID, workspaceID resource.TfeID) error {
	_, err := db.Exec(ctx, `
INSERT INTO agent_pool_allowed_workspaces (
    agent_pool_id,
    workspace_id
) VALUES (
    $1,
    $2
)`, poolID, workspaceID)
	return err
}

func (db *db) deleteAgentPoolAllowedWorkspace(ctx context.Context, poolID, workspaceID resource.TfeID) error {
	_, err := db.Exec(ctx, `
DELETE
FROM agent_pool_allowed_workspaces
WHERE agent_pool_id = $1
AND workspace_id = $2
`, poolID, workspaceID)
	return err
}

func (db *db) getPool(ctx context.Context, poolID resource.TfeID) (*Pool, error) {
	rows := db.Query(ctx, `
SELECT ap.agent_pool_id, ap.name, ap.created_at, ap.organization_name, ap.organization_scoped,
    (
        SELECT array_agg(w.workspace_id)::text[]
        FROM workspaces w
        WHERE w.agent_pool_id = ap.agent_pool_id
    ) AS workspace_ids,
    (
        SELECT array_agg(aw.workspace_id)::text[]
        FROM agent_pool_allowed_workspaces aw
        WHERE aw.agent_pool_id = ap.agent_pool_id
    ) AS allowed_workspace_ids
FROM agent_pools ap
WHERE ap.agent_pool_id = $1
GROUP BY ap.agent_pool_id
`, poolID)
	return sql.CollectOneRow(rows, scanAgentPool)
}

func (db *db) getPoolByTokenID(ctx context.Context, tokenID resource.TfeID) (*Pool, error) {
	rows := db.Query(ctx, `
SELECT ap.agent_pool_id, ap.name, ap.created_at, ap.organization_name, ap.organization_scoped,
    (
        SELECT array_agg(w.workspace_id)::text[]
        FROM workspaces w
        WHERE w.agent_pool_id = ap.agent_pool_id
    ) AS workspace_ids,
    (
        SELECT array_agg(aw.workspace_id)::text[]
        FROM agent_pool_allowed_workspaces aw
        WHERE aw.agent_pool_id = ap.agent_pool_id
    ) AS allowed_workspace_ids
FROM agent_pools ap
JOIN agent_tokens at USING (agent_pool_id)
WHERE at.agent_token_id = $1
GROUP BY ap.agent_pool_id
`, tokenID)
	return sql.CollectOneRow(rows, scanAgentPool)
}

func (db *db) listPoolsByOrganization(ctx context.Context, organization organization.Name, opts listPoolOptions) ([]*Pool, error) {
	rows := db.Query(ctx, `
SELECT ap.agent_pool_id, ap.name, ap.created_at, ap.organization_name, ap.organization_scoped,
    (
        SELECT array_agg(w.workspace_id)::text[]
        FROM workspaces w
        WHERE w.agent_pool_id = ap.agent_pool_id
    ) AS workspace_ids,
    (
        SELECT array_agg(aw.workspace_id)::text[]
        FROM agent_pool_allowed_workspaces aw
        WHERE aw.agent_pool_id = ap.agent_pool_id
    ) AS allowed_workspace_ids
FROM agent_pools ap
LEFT JOIN (agent_pool_allowed_workspaces aw JOIN workspaces w USING (workspace_id)) ON ap.agent_pool_id = aw.agent_pool_id
WHERE ap.organization_name = $1
AND   (($2::text IS NULL) OR ap.name LIKE '%' || $2 || '%')
AND   (($3::text IS NULL) OR
       ap.organization_scoped OR
       w.name = $3
      )
AND   (($4::text IS NULL) OR
       ap.organization_scoped OR
       w.workspace_id = $4
      )
GROUP BY ap.agent_pool_id
ORDER BY ap.created_at DESC
`,
		organization,
		opts.NameSubstring,
		opts.AllowedWorkspaceName,
		opts.AllowedWorkspaceID,
	)
	return sql.CollectRows(rows, scanAgentPool)
}

func (db *db) deleteAgentPool(ctx context.Context, poolID resource.TfeID) error {
	_, err := db.Exec(ctx, `
DELETE
FROM agent_pools
WHERE agent_pool_id = $1
RETURNING agent_pool_id, name, created_at, organization_name, organization_scoped
`, poolID)
	return err
}

func scanAgentPool(row pgx.CollectableRow) (*Pool, error) {
	pool, err := pgx.RowToAddrOfStructByName[Pool](row)
	if err != nil {
		return nil, err
	}
	pool.CreatedAt = pool.CreatedAt.UTC()
	return pool, nil
}
