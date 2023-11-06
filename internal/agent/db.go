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
	OrganizationScoped  bool               `json:"organization_scoped"`
	WorkspaceIds        []string           `json:"workspace_ids"`
	AllowedWorkspaceIds []string           `json:"allowed_workspace_ids"`
}

func (r poolresult) toPool() *Pool {
	return &Pool{
		ID:                 r.AgentPoolID.String,
		Name:               r.Name.String,
		CreatedAt:          r.CreatedAt.Time.UTC(),
		Organization:       r.OrganizationName.String,
		OrganizationScoped: r.OrganizationScoped,
		Workspaces:         r.WorkspaceIds,
		AllowedWorkspaces:  r.AllowedWorkspaceIds,
	}
}

// agentresult is the result of a database query for an agent
type agentresult struct {
	AgentID      pgtype.Text        `json:"agent_id"`
	Name         pgtype.Text        `json:"name"`
	Concurrency  pgtype.Int4        `json:"concurrency"`
	Server       bool               `json:"server"`
	IPAddress    pgtype.Inet        `json:"ip_address"`
	LastPingAt   pgtype.Timestamptz `json:"last_ping_at"`
	Status       pgtype.Text        `json:"status"`
	AgentTokenID pgtype.Text        `json:"agent_token_id"`
}

func (r agentresult) toAgent() *Agent {
	agent := &Agent{
		ID:          r.AgentID.String,
		Concurrency: int(r.Concurrency.Int),
		Server:      r.Server,
		IPAddress:   r.IPAddress.IPNet.IP,
		LastPingAt:  r.LastPingAt.Time.UTC(),
		Status:      AgentStatus(r.Status.String),
	}
	if r.Name.Status == pgtype.Present {
		agent.Name = &r.Name.String
	}
	if r.AgentTokenID.Status == pgtype.Present {
		agent.AgentPoolID = &r.AgentTokenID.String
	}
	return agent
}

// jobresult is the result of a database query for an job
type jobresult struct {
	RunID         pgtype.Text `json:"run_id"`
	Phase         pgtype.Text `json:"phase"`
	Status        pgtype.Text `json:"status"`
	ExecutionMode pgtype.Text `json:"execution_mode"`
	WorkspaceID   pgtype.Text `json:"workspace_id"`
	AgentID       pgtype.Text `json:"agent_id"`
}

func (r jobresult) toJob() *Job {
	job := &Job{
		JobSpec: JobSpec{
			RunID: r.RunID.String,
			Phase: internal.PhaseType(r.Phase.String),
		},
		Status:        JobStatus(r.Status.String),
		ExecutionMode: workspace.ExecutionMode(r.ExecutionMode.String),
		WorkspaceID:   r.WorkspaceID.String,
	}
	if r.AgentID.Status == pgtype.Present {
		job.AgentID = &r.AgentID.String
	}
	return job
}

type db struct {
	*sql.DB
}

func (db *db) createAgent(ctx context.Context, agent *Agent) error {
	_, err := db.Conn(ctx).InsertAgent(ctx, pggen.InsertAgentParams{
		AgentID:      sql.String(agent.ID),
		Name:         sql.StringPtr(agent.Name),
		Concurrency:  sql.Int4(agent.Concurrency),
		IPAddress:    sql.Inet(agent.IPAddress),
		Status:       sql.String(string(agent.Status)),
		AgentTokenID: sql.StringPtr(agent.AgentPoolID),
	})
	return err
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

func (db *db) createPool(ctx context.Context, pool *Pool) error {
	err := db.Tx(ctx, func(ctx context.Context, q pggen.Querier) error {
		_, err := db.Conn(ctx).InsertAgentPool(ctx, pggen.InsertAgentPoolParams{
			AgentPoolID:        sql.String(pool.ID),
			Name:               sql.String(pool.Name),
			CreatedAt:          sql.Timestamptz(pool.CreatedAt),
			OrganizationName:   sql.String(pool.Organization),
			OrganizationScoped: pool.OrganizationScoped,
		})
		if err != nil {
			return err
		}
		for _, workspaceID := range pool.AllowedWorkspaces {
			_, err := q.InsertAgentPoolAllowedWorkspaces(ctx, sql.String(pool.ID), sql.String(workspaceID))
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
	// TODO: remove tx or lock table
	err := db.Tx(ctx, func(ctx context.Context, q pggen.Querier) error {
		_, err := q.UpdateAgentPool(ctx, pggen.UpdateAgentPoolParams{
			PoolID:             sql.String(pool.ID),
			Name:               sql.String(pool.Name),
			OrganizationScoped: pool.OrganizationScoped,
		})
		if err != nil {
			return err
		}
		// Rather than work out which allowed workspaces have been added/removed,
		// instead simply delete all and re-insert.
		_, err = q.DeleteAgentPoolAllowedWorkspaces(ctx, sql.String(pool.ID))
		if err != nil {
			return err
		}
		for _, workspaceID := range pool.AllowedWorkspaces {
			_, err := q.InsertAgentPoolAllowedWorkspaces(ctx, sql.String(pool.ID), sql.String(workspaceID))
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return sql.Error(err)
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

func (db *db) listPools(ctx context.Context, opts listPoolOptions) ([]*Pool, error) {
	params := pggen.FindAgentPoolsParams{
		OrganizationName:     sql.StringPtr(opts.Organization),
		NameSubstring:        sql.StringPtr(opts.NameSubstring),
		AllowedWorkspaceName: sql.StringPtr(opts.AllowedWorkspaceName),
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

func (db *db) deletePool(ctx context.Context, poolID string) (organization string, err error) {
	result, err := db.Conn(ctx).DeleteAgentPool(ctx, sql.String(poolID))
	if err != nil {
		return "", sql.Error(err)
	}
	return result.String, nil
}

func (db *db) createJob(ctx context.Context, job *Job) error {
	_, err := db.Conn(ctx).InsertJob(ctx, pggen.InsertJobParams{
		RunID:  sql.String(job.RunID),
		Phase:  sql.String(string(job.Phase)),
		Status: sql.String(string(job.Status)),
	})
	return sql.Error(err)
}

func (db *db) allocateJob(ctx context.Context, spec JobSpec, agentID string) error {
	_, err := db.Conn(ctx).AllocateJob(ctx, pggen.AllocateJobParams{
		AgentID: sql.String(agentID),
		RunID:   sql.String(spec.RunID),
		Phase:   sql.String(string(spec.Phase)),
	})
	return sql.Error(err)
}

func (db *db) getAllocatedJobs(ctx context.Context, agentID string) ([]*Job, error) {
	rows, err := db.Conn(ctx).FindAllocatedJobs(ctx, sql.String(agentID))
	if err != nil {
		return nil, sql.Error(err)
	}
	jobs := make([]*Job, len(rows))
	for i, r := range rows {
		jobs[i] = jobresult(r).toJob()
	}
	return jobs, nil
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

func (db *db) updateJobStatus(ctx context.Context, spec JobSpec, status JobStatus) error {
	_, err := db.Conn(ctx).UpdateJobStatus(ctx, pggen.UpdateJobStatusParams{
		RunID:  sql.String(spec.RunID),
		Phase:  sql.String(string(spec.Phase)),
		Status: sql.String(string(status)),
	})
	return sql.Error(err)
}
