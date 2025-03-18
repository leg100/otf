package runner

import (
	"context"
	"net/netip"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/sql"
)

var q = &Queries{}

type db struct {
	*sql.DB
}

// runners

type runnerMetaResult struct {
	RunnerID     resource.ID
	Name         pgtype.Text
	Version      pgtype.Text
	MaxJobs      pgtype.Int4
	IPAddress    netip.Addr
	LastPingAt   pgtype.Timestamptz
	LastStatusAt pgtype.Timestamptz
	Status       pgtype.Text
	AgentPoolID  *resource.ID
	AgentPool    *AgentPool
	CurrentJobs  int64
}

func (r runnerMetaResult) toRunnerMeta() *RunnerMeta {
	meta := &RunnerMeta{
		ID:           r.RunnerID,
		Name:         r.Name.String,
		Version:      r.Version.String,
		MaxJobs:      int(r.MaxJobs.Int32),
		CurrentJobs:  int(r.CurrentJobs),
		IPAddress:    r.IPAddress,
		LastPingAt:   r.LastPingAt.Time.UTC(),
		LastStatusAt: r.LastStatusAt.Time.UTC(),
		Status:       RunnerStatus(r.Status.String),
	}
	if r.AgentPool != nil {
		meta.AgentPool = &RunnerMetaAgentPool{
			ID:               r.AgentPool.AgentPoolID,
			Name:             r.AgentPool.Name.String,
			OrganizationName: r.AgentPool.OrganizationName.String,
		}
	}
	return meta
}

func (db *db) create(ctx context.Context, meta *RunnerMeta) error {
	params := sqlc.InsertRunnerParams{
		RunnerID:     meta.ID,
		Name:         sql.String(meta.Name),
		Version:      sql.String(meta.Version),
		MaxJobs:      sql.Int4(meta.MaxJobs),
		IPAddress:    meta.IPAddress,
		Status:       sql.String(string(meta.Status)),
		LastPingAt:   sql.Timestamptz(meta.LastPingAt),
		LastStatusAt: sql.Timestamptz(meta.LastStatusAt),
	}
	if meta.AgentPool != nil {
		params.AgentPoolID = &meta.AgentPool.ID
	}
	return db.Querier(ctx).InsertRunner(ctx, params)
}

func (db *db) update(ctx context.Context, runnerID resource.ID, fn func(context.Context, *RunnerMeta) error) error {
	_, err := sql.Updater(
		ctx,
		db.DB,
		func(ctx context.Context, q *sqlc.Queries) (*RunnerMeta, error) {
			result, err := q.FindRunnerByIDForUpdate(ctx, runnerID)
			if err != nil {
				return nil, err
			}
			return runnerMetaResult(result).toRunnerMeta(), nil
		},
		fn,
		func(ctx context.Context, q *sqlc.Queries, agent *RunnerMeta) error {
			_, err := q.UpdateRunner(ctx, sqlc.UpdateRunnerParams{
				RunnerID:     agent.ID,
				Status:       sql.String(string(agent.Status)),
				LastPingAt:   sql.Timestamptz(agent.LastPingAt),
				LastStatusAt: sql.Timestamptz(agent.LastStatusAt),
			})
			return err
		},
	)
	return err
}

func (db *db) get(ctx context.Context, runnerID resource.ID) (*RunnerMeta, error) {
	result, err := db.Querier(ctx).FindRunnerByID(ctx, runnerID)
	if err != nil {
		return nil, sql.Error(err)
	}
	return runnerMetaResult(result).toRunnerMeta(), nil
}

func (db *db) list(ctx context.Context) ([]*RunnerMeta, error) {
	rows, err := db.Querier(ctx).FindRunners(ctx)
	if err != nil {
		return nil, sql.Error(err)
	}
	agents := make([]*RunnerMeta, len(rows))
	for i, r := range rows {
		agents[i] = runnerMetaResult(r).toRunnerMeta()
	}
	return agents, nil
}

func (db *db) listServerRunners(ctx context.Context) ([]*RunnerMeta, error) {
	rows, err := db.Querier(ctx).FindServerRunners(ctx)
	if err != nil {
		return nil, sql.Error(err)
	}
	agents := make([]*RunnerMeta, len(rows))
	for i, r := range rows {
		agents[i] = runnerMetaResult(r).toRunnerMeta()
	}
	return agents, nil
}

func (db *db) listRunnersByOrganization(ctx context.Context, organization string) ([]*RunnerMeta, error) {
	rows, err := db.Querier(ctx).FindRunnersByOrganization(ctx, sql.String(organization))
	if err != nil {
		return nil, sql.Error(err)
	}
	agents := make([]*RunnerMeta, len(rows))
	for i, r := range rows {
		agents[i] = runnerMetaResult(r).toRunnerMeta()
	}
	return agents, nil
}

func (db *db) listRunnersByPool(ctx context.Context, poolID resource.ID) ([]*RunnerMeta, error) {
	rows, err := db.Querier(ctx).FindRunnersByPoolID(ctx, poolID)
	if err != nil {
		return nil, sql.Error(err)
	}
	runners := make([]*RunnerMeta, len(rows))
	for i, r := range rows {
		runners[i] = runnerMetaResult(r).toRunnerMeta()
	}
	return runners, nil
}

func (db *db) deleteRunner(ctx context.Context, runnerID resource.ID) error {
	_, err := db.Querier(ctx).DeleteRunner(ctx, runnerID)
	return sql.Error(err)
}

// jobs

type jobResult struct {
	JobID            resource.ID
	RunID            resource.ID
	Phase            pgtype.Text
	Status           pgtype.Text
	Signaled         pgtype.Bool
	RunnerID         *resource.ID
	AgentPoolID      *resource.ID
	WorkspaceID      resource.ID
	OrganizationName pgtype.Text
}

func (r jobResult) toJob() *Job {
	job := &Job{
		ID:           r.JobID,
		RunID:        r.RunID,
		Phase:        internal.PhaseType(r.Phase.String),
		Status:       JobStatus(r.Status.String),
		WorkspaceID:  r.WorkspaceID,
		Organization: r.OrganizationName.String,
		RunnerID:     r.RunnerID,
		AgentPoolID:  r.AgentPoolID,
	}
	if r.Signaled.Valid {
		job.Signaled = &r.Signaled.Bool
	}
	return job
}

func (db *db) createJob(ctx context.Context, job *Job) error {
	err := db.Querier(ctx).InsertJob(ctx, sqlc.InsertJobParams{
		JobID:  job.ID,
		RunID:  job.RunID,
		Phase:  sql.String(string(job.Phase)),
		Status: sql.String(string(job.Status)),
	})
	return sql.Error(err)
}

func (db *db) getAllocatedAndSignaledJobs(ctx context.Context, runnerID resource.ID) ([]*Job, error) {
	allocated, err := db.Querier(ctx).FindAllocatedJobs(ctx, &runnerID)
	if err != nil {
		return nil, sql.Error(err)
	}
	signaled, err := db.Querier(ctx).FindAndUpdateSignaledJobs(ctx, &runnerID)
	if err != nil {
		return nil, sql.Error(err)
	}
	jobs := make([]*Job, len(allocated)+len(signaled))
	for i, r := range allocated {
		jobs[i] = jobResult(r).toJob()
	}
	for i, r := range signaled {
		jobs[len(allocated)+i] = jobResult(r).toJob()
	}
	return jobs, nil
}

func (db *db) getJob(ctx context.Context, jobID resource.ID) (*Job, error) {
	result, err := db.Querier(ctx).FindJob(ctx, jobID)
	if err != nil {
		return nil, sql.Error(err)
	}
	return jobResult(result).toJob(), nil
}

func (db *db) listJobs(ctx context.Context) ([]*Job, error) {
	rows, err := db.Querier(ctx).FindJobs(ctx)
	if err != nil {
		return nil, sql.Error(err)
	}
	jobs := make([]*Job, len(rows))
	for i, r := range rows {
		jobs[i] = jobResult(r).toJob()
	}
	return jobs, nil
}

func (db *db) updateJob(ctx context.Context, jobID resource.ID, fn func(context.Context, *Job) error) (*Job, error) {
	return sql.Updater(
		ctx,
		db.DB,
		func(ctx context.Context, q *sqlc.Queries) (*Job, error) {
			result, err := q.FindJobForUpdate(ctx, jobID)
			if err != nil {
				return nil, err
			}
			return jobResult(result).toJob(), nil
		},
		fn,
		func(ctx context.Context, q *sqlc.Queries, job *Job) error {
			_, err := q.UpdateJob(ctx, sqlc.UpdateJobParams{
				Status:   sql.String(string(job.Status)),
				Signaled: sql.BoolPtr(job.Signaled),
				RunnerID: job.RunnerID,
				JobID:    job.ID,
			})
			return err
		},
	)
}

func (db *db) updateUnfinishedJobByRunID(ctx context.Context, runID resource.ID, fn func(context.Context, *Job) error) (*Job, error) {
	return sql.Updater(
		ctx,
		db.DB,
		func(ctx context.Context, q *sqlc.Queries) (*Job, error) {
			result, err := q.FindUnfinishedJobForUpdateByRunID(ctx, runID)
			if err != nil {
				return nil, err
			}
			return jobResult(result).toJob(), nil
		},
		fn,
		func(ctx context.Context, q *sqlc.Queries, job *Job) error {
			_, err := q.UpdateJob(ctx, sqlc.UpdateJobParams{
				Status:   sql.String(string(job.Status)),
				Signaled: sql.BoolPtr(job.Signaled),
				RunnerID: job.RunnerID,
				JobID:    job.ID,
			})
			return err
		},
	)
}

// agent tokens

type agentTokenRow struct {
	AgentTokenID resource.ID        `json:"agent_token_id"`
	CreatedAt    pgtype.Timestamptz `json:"created_at"`
	Description  pgtype.Text        `json:"description"`
	AgentPoolID  resource.ID        `json:"agent_pool_id"`
}

func (row agentTokenRow) toAgentToken() *agentToken {
	return &agentToken{
		ID:          row.AgentTokenID,
		CreatedAt:   row.CreatedAt.Time.UTC(),
		Description: row.Description.String,
		AgentPoolID: row.AgentPoolID,
	}
}

func (db *db) createAgentToken(ctx context.Context, token *agentToken) error {
	return db.Querier(ctx).InsertAgentToken(ctx, sqlc.InsertAgentTokenParams{
		AgentTokenID: token.ID,
		Description:  sql.String(token.Description),
		AgentPoolID:  token.AgentPoolID,
		CreatedAt:    sql.Timestamptz(token.CreatedAt.UTC()),
	})
}

func (db *db) getAgentTokenByID(ctx context.Context, id resource.ID) (*agentToken, error) {
	r, err := db.Querier(ctx).FindAgentTokenByID(ctx, id)
	if err != nil {
		return nil, sql.Error(err)
	}
	return agentTokenRow(r).toAgentToken(), nil
}

func (db *db) listAgentTokens(ctx context.Context, poolID resource.ID) ([]*agentToken, error) {
	rows, err := db.Querier(ctx).FindAgentTokensByAgentPoolID(ctx, poolID)
	if err != nil {
		return nil, sql.Error(err)
	}
	tokens := make([]*agentToken, len(rows))
	for i, r := range rows {
		tokens[i] = agentTokenRow(r).toAgentToken()
	}
	return tokens, nil
}

func (db *db) deleteAgentToken(ctx context.Context, id resource.ID) error {
	_, err := db.Querier(ctx).DeleteAgentTokenByID(ctx, id)
	if err != nil {
		return sql.Error(err)
	}
	return nil
}

// agent pools

type poolresult struct {
	AgentPoolID         resource.ID
	Name                pgtype.Text
	CreatedAt           pgtype.Timestamptz
	OrganizationName    pgtype.Text
	OrganizationScoped  pgtype.Bool
	WorkspaceIds        []pgtype.Text
	AllowedWorkspaceIds []pgtype.Text
}

func (r poolresult) toPool() (*Pool, error) {
	pool := &Pool{
		ID:                 r.AgentPoolID,
		Name:               r.Name.String,
		CreatedAt:          r.CreatedAt.Time.UTC(),
		Organization:       r.OrganizationName.String,
		OrganizationScoped: r.OrganizationScoped.Bool,
	}
	pool.AssignedWorkspaces = make([]resource.ID, len(r.WorkspaceIds))
	for i, wid := range r.WorkspaceIds {
		var err error
		pool.AssignedWorkspaces[i], err = resource.ParseID(wid.String)
		if err != nil {
			return nil, err
		}
	}
	pool.AllowedWorkspaces = make([]resource.ID, len(r.AllowedWorkspaceIds))
	for i, wid := range r.AllowedWorkspaceIds {
		var err error
		pool.AllowedWorkspaces[i], err = resource.ParseID(wid.String)
		if err != nil {
			return nil, err
		}
	}
	return pool, nil
}

func (db *db) createPool(ctx context.Context, pool *Pool) error {
	err := db.Tx(ctx, func(ctx context.Context, q *sqlc.Queries) error {
		err := db.Querier(ctx).InsertAgentPool(ctx, sqlc.InsertAgentPoolParams{
			AgentPoolID:        pool.ID,
			Name:               sql.String(pool.Name),
			CreatedAt:          sql.Timestamptz(pool.CreatedAt),
			OrganizationName:   sql.String(pool.Organization),
			OrganizationScoped: sql.Bool(pool.OrganizationScoped),
		})
		if err != nil {
			return err
		}
		for _, workspaceID := range pool.AllowedWorkspaces {
			err := q.InsertAgentPoolAllowedWorkspace(ctx, sqlc.InsertAgentPoolAllowedWorkspaceParams{
				PoolID:      pool.ID,
				WorkspaceID: workspaceID,
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (db *db) updatePool(ctx context.Context, pool *Pool) error {
	_, err := db.Querier(ctx).UpdateAgentPool(ctx, sqlc.UpdateAgentPoolParams{
		PoolID:             pool.ID,
		Name:               sql.String(pool.Name),
		OrganizationScoped: sql.Bool(pool.OrganizationScoped),
	})
	if err != nil {
		return sql.Error(err)
	}
	return nil
}

func (db *db) addAgentPoolAllowedWorkspace(ctx context.Context, poolID, workspaceID resource.ID) error {
	err := db.Querier(ctx).InsertAgentPoolAllowedWorkspace(ctx, sqlc.InsertAgentPoolAllowedWorkspaceParams{
		PoolID:      poolID,
		WorkspaceID: workspaceID,
	})
	if err != nil {
		return err
	}
	return nil
}

func (db *db) deleteAgentPoolAllowedWorkspace(ctx context.Context, poolID, workspaceID resource.ID) error {
	err := db.Querier(ctx).DeleteAgentPoolAllowedWorkspace(ctx, sqlc.DeleteAgentPoolAllowedWorkspaceParams{
		PoolID:      poolID,
		WorkspaceID: workspaceID,
	})
	if err != nil {
		return err
	}
	return nil
}

func (db *db) getPool(ctx context.Context, poolID resource.ID) (*Pool, error) {
	result, err := db.Querier(ctx).FindAgentPool(ctx, poolID)
	if err != nil {
		return nil, sql.Error(err)
	}
	return poolresult(result).toPool()
}

func (db *db) getPoolByTokenID(ctx context.Context, tokenID resource.ID) (*Pool, error) {
	result, err := db.Querier(ctx).FindAgentPoolByAgentTokenID(ctx, tokenID)
	if err != nil {
		return nil, sql.Error(err)
	}
	return poolresult(result).toPool()
}

func (db *db) listPoolsByOrganization(ctx context.Context, organization string, opts listPoolOptions) ([]*Pool, error) {
	rows, err := db.Querier(ctx).FindAgentPoolsByOrganization(ctx, sqlc.FindAgentPoolsByOrganizationParams{
		OrganizationName:     sql.String(organization),
		NameSubstring:        sql.StringPtr(opts.NameSubstring),
		AllowedWorkspaceName: sql.StringPtr(opts.AllowedWorkspaceName),
		AllowedWorkspaceID:   sql.IDPtr(opts.AllowedWorkspaceID),
	})
	if err != nil {
		return nil, sql.Error(err)
	}
	pools := make([]*Pool, len(rows))
	for i, r := range rows {
		var err error
		pools[i], err = poolresult(r).toPool()
		if err != nil {
			return nil, err
		}
	}
	return pools, nil
}

func (db *db) deleteAgentPool(ctx context.Context, poolID resource.ID) error {
	_, err := db.Querier(ctx).DeleteAgentPool(ctx, poolID)
	if err != nil {
		return sql.Error(err)
	}
	return nil
}
