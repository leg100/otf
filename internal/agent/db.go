package agent

import (
	"context"

	"github.com/jackc/pgtype"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/sql/pggen"
	"github.com/leg100/otf/internal/workspace"
)

// poolresult is the result of a database query for an agent pool
type poolresult struct {
	AgentPoolID         pgtype.Text        `json:"agent_pool_id"`
	Name                pgtype.Text        `json:"name"`
	CreatedAt           pgtype.Timestamptz `json:"created_at"`
	OrganizationName    pgtype.Text        `json:"organization_name"`
	OrganizationScoped  pgtype.Bool        `json:"organization_scoped"`
	WorkspaceIds        []string           `json:"workspace_ids"`
	AllowedWorkspaceIds []string           `json:"allowed_workspace_ids"`
}

func (r poolresult) toPool() *Pool {
	return &Pool{
		ID:                 r.AgentPoolID.String,
		Name:               r.Name.String,
		CreatedAt:          r.CreatedAt.Time.UTC(),
		Organization:       r.OrganizationName.String,
		OrganizationScoped: r.OrganizationScoped.Bool,
		AssignedWorkspaces: r.WorkspaceIds,
		AllowedWorkspaces:  r.AllowedWorkspaceIds,
	}
}

// agentresult is the result of a database query for an agent
type agentresult struct {
	AgentID      pgtype.Text        `json:"agent_id"`
	Name         pgtype.Text        `json:"name"`
	Version      pgtype.Text        `json:"version"`
	Concurrency  pgtype.Int4        `json:"concurrency"`
	IPAddress    pgtype.Inet        `json:"ip_address"`
	LastPingAt   pgtype.Timestamptz `json:"last_ping_at"`
	LastStatusAt pgtype.Timestamptz `json:"last_status_at"`
	Status       pgtype.Text        `json:"status"`
	AgentPoolID  pgtype.Text        `json:"agent_pool_id"`
}

func (r agentresult) toAgent() *Agent {
	agent := &Agent{
		ID:           r.AgentID.String,
		Version:      r.Version.String,
		Concurrency:  int(r.Concurrency.Int),
		IPAddress:    r.IPAddress.IPNet.IP,
		LastPingAt:   r.LastPingAt.Time.UTC(),
		LastStatusAt: r.LastStatusAt.Time.UTC(),
		Status:       AgentStatus(r.Status.String),
	}
	if r.Name.Status == pgtype.Present {
		agent.Name = &r.Name.String
	}
	if r.AgentPoolID.Status == pgtype.Present {
		agent.AgentPoolID = &r.AgentPoolID.String
	}
	return agent
}

// jobresult is the result of a database query for an job
type jobresult struct {
	RunID            pgtype.Text `json:"run_id"`
	Phase            pgtype.Text `json:"phase"`
	Status           pgtype.Text `json:"status"`
	Signaled         pgtype.Bool `json:"signaled"`
	AgentID          pgtype.Text `json:"agent_id"`
	ExecutionMode    pgtype.Text `json:"execution_mode"`
	WorkspaceID      pgtype.Text `json:"workspace_id"`
	OrganizationName pgtype.Text `json:"organization_name"`
}

func (r jobresult) toJob() *Job {
	job := &Job{
		Spec: JobSpec{
			RunID: r.RunID.String,
			Phase: internal.PhaseType(r.Phase.String),
		},
		Status:        JobStatus(r.Status.String),
		ExecutionMode: workspace.ExecutionMode(r.ExecutionMode.String),
		WorkspaceID:   r.WorkspaceID.String,
		Organization:  r.OrganizationName.String,
	}
	if r.AgentID.Status == pgtype.Present {
		job.AgentID = &r.AgentID.String
	}
	if r.Signaled.Status == pgtype.Present {
		job.Signaled = &r.Signaled.Bool
	}
	return job
}

type agentTokenRow struct {
	AgentTokenID pgtype.Text        `json:"agent_token_id"`
	CreatedAt    pgtype.Timestamptz `json:"created_at"`
	Description  pgtype.Text        `json:"description"`
	AgentPoolID  pgtype.Text        `json:"agent_pool_id"`
}

func (row agentTokenRow) toAgentToken() *agentToken {
	return &agentToken{
		ID:          row.AgentTokenID.String,
		CreatedAt:   row.CreatedAt.Time.UTC(),
		Description: row.Description.String,
		AgentPoolID: row.AgentPoolID.String,
	}
}

type db struct {
	*sql.DB
}

// pools

func (db *db) createPool(ctx context.Context, pool *Pool) error {
	err := db.Tx(ctx, func(ctx context.Context, q pggen.Querier) error {
		_, err := db.Conn(ctx).InsertAgentPool(ctx, pggen.InsertAgentPoolParams{
			AgentPoolID:        sql.String(pool.ID),
			Name:               sql.String(pool.Name),
			CreatedAt:          sql.Timestamptz(pool.CreatedAt),
			OrganizationName:   sql.String(pool.Organization),
			OrganizationScoped: sql.Bool(pool.OrganizationScoped),
		})
		if err != nil {
			return err
		}
		for _, workspaceID := range pool.AllowedWorkspaces {
			_, err := q.InsertAgentPoolAllowedWorkspace(ctx, sql.String(pool.ID), sql.String(workspaceID))
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
	_, err := db.Conn(ctx).UpdateAgentPool(ctx, pggen.UpdateAgentPoolParams{
		PoolID:             sql.String(pool.ID),
		Name:               sql.String(pool.Name),
		OrganizationScoped: sql.Bool(pool.OrganizationScoped),
	})
	if err != nil {
		return sql.Error(err)
	}
	return nil
}

func (db *db) addAgentPoolAllowedWorkspace(ctx context.Context, poolID, workspaceID string) error {
	_, err := db.Conn(ctx).InsertAgentPoolAllowedWorkspace(ctx, sql.String(poolID), sql.String(workspaceID))
	if err != nil {
		return err
	}
	return nil
}

func (db *db) deleteAgentPoolAllowedWorkspace(ctx context.Context, poolID, workspaceID string) error {
	_, err := db.Conn(ctx).DeleteAgentPoolAllowedWorkspace(ctx, sql.String(poolID), sql.String(workspaceID))
	if err != nil {
		return err
	}
	return nil
}

func (db *db) getPool(ctx context.Context, poolID string) (*Pool, error) {
	result, err := db.Conn(ctx).FindAgentPool(ctx, sql.String(poolID))
	if err != nil {
		return nil, sql.Error(err)
	}
	return poolresult(result).toPool(), nil
}

func (db *db) getPoolByTokenID(ctx context.Context, tokenID string) (*Pool, error) {
	result, err := db.Conn(ctx).FindAgentPoolByAgentTokenID(ctx, sql.String(tokenID))
	if err != nil {
		return nil, sql.Error(err)
	}
	return poolresult(result).toPool(), nil
}

func (db *db) listPools(ctx context.Context, opts listPoolOptions) ([]*Pool, error) {
	params := pggen.FindAgentPoolsParams{
		OrganizationName:     sql.StringPtr(opts.Organization),
		NameSubstring:        sql.StringPtr(opts.NameSubstring),
		AllowedWorkspaceName: sql.StringPtr(opts.AllowedWorkspaceName),
		AllowedWorkspaceID:   sql.StringPtr(opts.AllowedWorkspaceID),
	}
	rows, err := db.Conn(ctx).FindAgentPools(ctx, params)
	if err != nil {
		return nil, sql.Error(err)
	}
	pools := make([]*Pool, len(rows))
	for i, r := range rows {
		pools[i] = poolresult(r).toPool()
	}
	return pools, nil
}

func (db *db) deleteAgentPool(ctx context.Context, poolID string) error {
	_, err := db.Conn(ctx).DeleteAgentPool(ctx, sql.String(poolID))
	if err != nil {
		return sql.Error(err)
	}
	return nil
}

// agents

func (db *db) createAgent(ctx context.Context, agent *Agent) error {
	_, err := db.Conn(ctx).InsertAgent(ctx, pggen.InsertAgentParams{
		AgentID:      sql.String(agent.ID),
		Name:         sql.StringPtr(agent.Name),
		Version:      sql.String(agent.Version),
		Concurrency:  sql.Int4(agent.Concurrency),
		IPAddress:    sql.Inet(agent.IPAddress),
		Status:       sql.String(string(agent.Status)),
		LastPingAt:   sql.Timestamptz(agent.LastPingAt),
		LastStatusAt: sql.Timestamptz(agent.LastStatusAt),
		AgentPoolID:  sql.StringPtr(agent.AgentPoolID),
	})
	return err
}

func (db *db) updateAgent(ctx context.Context, agentID string, fn func(*Agent) error) error {
	err := db.Tx(ctx, func(ctx context.Context, q pggen.Querier) error {
		result, err := q.FindAgentByIDForUpdate(ctx, sql.String(agentID))
		if err != nil {
			return err
		}
		agent := agentresult(result).toAgent()
		if err := fn(agent); err != nil {
			return err
		}
		_, err = q.UpdateAgent(ctx, pggen.UpdateAgentParams{
			AgentID:      sql.String(agent.ID),
			Status:       sql.String(string(agent.Status)),
			LastPingAt:   sql.Timestamptz(agent.LastPingAt),
			LastStatusAt: sql.Timestamptz(agent.LastStatusAt),
		})
		return err
	})
	if err != nil {
		return sql.Error(err)
	}
	return nil
}

func (db *db) getAgent(ctx context.Context, agentID string) (*Agent, error) {
	result, err := db.Conn(ctx).FindAgentByID(ctx, sql.String(agentID))
	if err != nil {
		return nil, sql.Error(err)
	}
	return agentresult(result).toAgent(), nil
}

func (db *db) listAgents(ctx context.Context) ([]*Agent, error) {
	rows, err := db.Conn(ctx).FindAgents(ctx)
	if err != nil {
		return nil, sql.Error(err)
	}
	agents := make([]*Agent, len(rows))
	for i, r := range rows {
		agents[i] = agentresult(r).toAgent()
	}
	return agents, nil
}

func (db *db) listServerAgents(ctx context.Context) ([]*Agent, error) {
	rows, err := db.Conn(ctx).FindServerAgents(ctx)
	if err != nil {
		return nil, sql.Error(err)
	}
	agents := make([]*Agent, len(rows))
	for i, r := range rows {
		agents[i] = agentresult(r).toAgent()
	}
	return agents, nil
}

func (db *db) listAgentsByOrganization(ctx context.Context, organization string) ([]*Agent, error) {
	rows, err := db.Conn(ctx).FindAgentsByOrganization(ctx, sql.String(organization))
	if err != nil {
		return nil, sql.Error(err)
	}
	agents := make([]*Agent, len(rows))
	for i, r := range rows {
		agents[i] = agentresult(r).toAgent()
	}
	return agents, nil
}

func (db *db) listAgentsByPool(ctx context.Context, poolID string) ([]*Agent, error) {
	rows, err := db.Conn(ctx).FindAgentsByPoolID(ctx, sql.String(poolID))
	if err != nil {
		return nil, sql.Error(err)
	}
	agents := make([]*Agent, len(rows))
	for i, r := range rows {
		agents[i] = agentresult(r).toAgent()
	}
	return agents, nil
}

func (db *db) deleteAgent(ctx context.Context, agentID string) error {
	_, err := db.Conn(ctx).DeleteAgent(ctx, sql.String(agentID))
	return sql.Error(err)
}

func (db *db) createJob(ctx context.Context, job *Job) error {
	_, err := db.Conn(ctx).InsertJob(ctx, pggen.InsertJobParams{
		RunID:  sql.String(job.Spec.RunID),
		Phase:  sql.String(string(job.Spec.Phase)),
		Status: sql.String(string(job.Status)),
	})
	return sql.Error(err)
}

func (db *db) getAllocatedAndSignaledJobs(ctx context.Context, agentID string) ([]*Job, error) {
	allocated, err := db.Conn(ctx).FindAllocatedJobs(ctx, sql.String(agentID))
	if err != nil {
		return nil, sql.Error(err)
	}
	signaled, err := db.Conn(ctx).FindAndUpdateSignaledJobs(ctx, sql.String(agentID))
	if err != nil {
		return nil, sql.Error(err)
	}
	jobs := make([]*Job, len(allocated)+len(signaled))
	for i, r := range allocated {
		jobs[i] = jobresult(r).toJob()
	}
	for i, r := range signaled {
		jobs[len(allocated)+i] = jobresult(r).toJob()
	}
	return jobs, nil
}

func (db *db) getJob(ctx context.Context, spec JobSpec) (*Job, error) {
	result, err := db.Conn(ctx).FindJob(ctx, sql.String(spec.RunID), sql.String(string(spec.Phase)))
	if err != nil {
		return nil, sql.Error(err)
	}
	return jobresult(result).toJob(), nil
}

func (db *db) listJobs(ctx context.Context) ([]*Job, error) {
	rows, err := db.Conn(ctx).FindJobs(ctx)
	if err != nil {
		return nil, sql.Error(err)
	}
	jobs := make([]*Job, len(rows))
	for i, r := range rows {
		jobs[i] = jobresult(r).toJob()
	}
	return jobs, nil
}

func (db *db) updateJob(ctx context.Context, spec JobSpec, fn func(*Job) error) (*Job, error) {
	var job *Job
	err := db.Tx(ctx, func(ctx context.Context, q pggen.Querier) error {
		result, err := q.FindJobForUpdate(ctx, sql.String(spec.RunID), sql.String(string(spec.Phase)))
		if err != nil {
			return err
		}
		job = jobresult(result).toJob()
		if err := fn(job); err != nil {
			return err
		}
		_, err = q.UpdateJob(ctx, pggen.UpdateJobParams{
			Status:   sql.String(string(job.Status)),
			Signaled: sql.BoolPtr(job.Signaled),
			AgentID:  sql.StringPtr(job.AgentID),
			RunID:    result.RunID,
			Phase:    result.Phase,
		})
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, sql.Error(err)
	}
	return job, nil
}

// agent tokens

func (db *db) createAgentToken(ctx context.Context, token *agentToken) error {
	_, err := db.Conn(ctx).InsertAgentToken(ctx, pggen.InsertAgentTokenParams{
		AgentTokenID: sql.String(token.ID),
		Description:  sql.String(token.Description),
		AgentPoolID:  sql.String(token.AgentPoolID),
		CreatedAt:    sql.Timestamptz(token.CreatedAt.UTC()),
	})
	return err
}

func (db *db) getAgentTokenByID(ctx context.Context, id string) (*agentToken, error) {
	r, err := db.Conn(ctx).FindAgentTokenByID(ctx, sql.String(id))
	if err != nil {
		return nil, sql.Error(err)
	}
	return agentTokenRow(r).toAgentToken(), nil
}

func (db *db) listAgentTokens(ctx context.Context, poolID string) ([]*agentToken, error) {
	rows, err := db.Conn(ctx).FindAgentTokensByAgentPoolID(ctx, sql.String(poolID))
	if err != nil {
		return nil, sql.Error(err)
	}
	tokens := make([]*agentToken, len(rows))
	for i, r := range rows {
		tokens[i] = agentTokenRow(r).toAgentToken()
	}
	return tokens, nil
}

func (db *db) deleteAgentToken(ctx context.Context, id string) error {
	_, err := db.Conn(ctx).DeleteAgentTokenByID(ctx, sql.String(id))
	if err != nil {
		return sql.Error(err)
	}
	return nil
}
