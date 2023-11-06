package agent

import (
	"context"

	"github.com/jackc/pgtype"
	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/sql/pggen"
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
	IpAddress    pgtype.Inet        `json:"ip_address"`
	LastPingAt   pgtype.Timestamptz `json:"last_ping_at"`
	Status       pgtype.Text        `json:"status"`
	AgentTokenID pgtype.Text        `json:"agent_token_id"`
}

func (r agentresult) toPool() *Agent {
	agent := &Agent{
		ID:          r.AgentID.String,
		Concurrency: int(r.Concurrency.Int),
		Server:      r.Server,
		IPAddress:   r.IpAddress.IPNet.IP,
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

type db struct {
	*sql.DB
}

func (db *db) createAgent(ctx context.Context, agent *Agent) error {
	_, err := db.Conn(ctx).InsertAgent(ctx, pggen.InsertAgentParams{
		AgentID:      sql.String(agent.ID),
		Name:         sql.StringPtr(agent.Name),
		Concurrency:  sql.Int4(agent.Concurrency),
		IpAddress:    sql.Inet(agent.IPAddress),
		Status:       sql.String(string(agent.Status)),
		AgentTokenID: sql.StringPtr(agent.AgentPoolID),
	})
	return err
}

func (db *db) listAgents(ctx context.Context) ([]*Agent, error) {
	return nil, nil
}

func (db *db) listAgentsByOrganization(ctx context.Context, organization string) ([]*Agent, error) {
	return nil, nil
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

func (db *db) reallocateJob(ctx context.Context, job *Job) error {
	return nil
}

func (db *db) getAllocatedJobs(ctx context.Context, agentID string) ([]*Job, error) {
	return nil, nil
}
