// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: agent.sql

package sqlc

import (
	"context"
	"net/netip"

	"github.com/jackc/pgx/v5/pgtype"
)

const deleteAgent = `-- name: DeleteAgent :one
DELETE
FROM agents
WHERE agent_id = $1
RETURNING agent_id, name, version, max_jobs, ip_address, last_ping_at, last_status_at, status, agent_pool_id
`

func (q *Queries) DeleteAgent(ctx context.Context, agentID pgtype.Text) (Agent, error) {
	row := q.db.QueryRow(ctx, deleteAgent, agentID)
	var i Agent
	err := row.Scan(
		&i.AgentID,
		&i.Name,
		&i.Version,
		&i.MaxJobs,
		&i.IPAddress,
		&i.LastPingAt,
		&i.LastStatusAt,
		&i.Status,
		&i.AgentPoolID,
	)
	return i, err
}

const findAgentByID = `-- name: FindAgentByID :one
SELECT
    a.agent_id, a.name, a.version, a.max_jobs, a.ip_address, a.last_ping_at, a.last_status_at, a.status, a.agent_pool_id,
    ( SELECT count(*)
      FROM jobs j
      WHERE a.agent_id = j.agent_id
      AND j.status IN ('allocated', 'running')
    ) AS current_jobs
FROM agents a
LEFT JOIN jobs j USING (agent_id)
WHERE a.agent_id = $1
GROUP BY a.agent_id
`

type FindAgentByIDRow struct {
	AgentID      pgtype.Text
	Name         pgtype.Text
	Version      pgtype.Text
	MaxJobs      pgtype.Int4
	IPAddress    netip.Addr
	LastPingAt   pgtype.Timestamptz
	LastStatusAt pgtype.Timestamptz
	Status       pgtype.Text
	AgentPoolID  pgtype.Text
	CurrentJobs  int64
}

func (q *Queries) FindAgentByID(ctx context.Context, agentID pgtype.Text) (FindAgentByIDRow, error) {
	row := q.db.QueryRow(ctx, findAgentByID, agentID)
	var i FindAgentByIDRow
	err := row.Scan(
		&i.AgentID,
		&i.Name,
		&i.Version,
		&i.MaxJobs,
		&i.IPAddress,
		&i.LastPingAt,
		&i.LastStatusAt,
		&i.Status,
		&i.AgentPoolID,
		&i.CurrentJobs,
	)
	return i, err
}

const findAgentByIDForUpdate = `-- name: FindAgentByIDForUpdate :one
SELECT
    a.agent_id, a.name, a.version, a.max_jobs, a.ip_address, a.last_ping_at, a.last_status_at, a.status, a.agent_pool_id,
    ( SELECT count(*)
      FROM jobs j
      WHERE a.agent_id = j.agent_id
      AND j.status IN ('allocated', 'running')
    ) AS current_jobs
FROM agents a
WHERE a.agent_id = $1
FOR UPDATE OF a
`

type FindAgentByIDForUpdateRow struct {
	AgentID      pgtype.Text
	Name         pgtype.Text
	Version      pgtype.Text
	MaxJobs      pgtype.Int4
	IPAddress    netip.Addr
	LastPingAt   pgtype.Timestamptz
	LastStatusAt pgtype.Timestamptz
	Status       pgtype.Text
	AgentPoolID  pgtype.Text
	CurrentJobs  int64
}

func (q *Queries) FindAgentByIDForUpdate(ctx context.Context, agentID pgtype.Text) (FindAgentByIDForUpdateRow, error) {
	row := q.db.QueryRow(ctx, findAgentByIDForUpdate, agentID)
	var i FindAgentByIDForUpdateRow
	err := row.Scan(
		&i.AgentID,
		&i.Name,
		&i.Version,
		&i.MaxJobs,
		&i.IPAddress,
		&i.LastPingAt,
		&i.LastStatusAt,
		&i.Status,
		&i.AgentPoolID,
		&i.CurrentJobs,
	)
	return i, err
}

const findAgents = `-- name: FindAgents :many
SELECT
    a.agent_id, a.name, a.version, a.max_jobs, a.ip_address, a.last_ping_at, a.last_status_at, a.status, a.agent_pool_id,
    ( SELECT count(*)
      FROM jobs j
      WHERE a.agent_id = j.agent_id
      AND j.status IN ('allocated', 'running')
    ) AS current_jobs
FROM agents a
GROUP BY a.agent_id
ORDER BY a.last_ping_at DESC
`

type FindAgentsRow struct {
	AgentID      pgtype.Text
	Name         pgtype.Text
	Version      pgtype.Text
	MaxJobs      pgtype.Int4
	IPAddress    netip.Addr
	LastPingAt   pgtype.Timestamptz
	LastStatusAt pgtype.Timestamptz
	Status       pgtype.Text
	AgentPoolID  pgtype.Text
	CurrentJobs  int64
}

func (q *Queries) FindAgents(ctx context.Context) ([]FindAgentsRow, error) {
	rows, err := q.db.Query(ctx, findAgents)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []FindAgentsRow
	for rows.Next() {
		var i FindAgentsRow
		if err := rows.Scan(
			&i.AgentID,
			&i.Name,
			&i.Version,
			&i.MaxJobs,
			&i.IPAddress,
			&i.LastPingAt,
			&i.LastStatusAt,
			&i.Status,
			&i.AgentPoolID,
			&i.CurrentJobs,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const findAgentsByOrganization = `-- name: FindAgentsByOrganization :many
SELECT
    a.agent_id, a.name, a.version, a.max_jobs, a.ip_address, a.last_ping_at, a.last_status_at, a.status, a.agent_pool_id,
    ( SELECT count(*)
      FROM jobs j
      WHERE a.agent_id = j.agent_id
      AND j.status IN ('allocated', 'running')
    ) AS current_jobs
FROM agents a
JOIN agent_pools ap USING (agent_pool_id)
WHERE ap.organization_name = $1
GROUP BY a.agent_id
ORDER BY last_ping_at DESC
`

type FindAgentsByOrganizationRow struct {
	AgentID      pgtype.Text
	Name         pgtype.Text
	Version      pgtype.Text
	MaxJobs      pgtype.Int4
	IPAddress    netip.Addr
	LastPingAt   pgtype.Timestamptz
	LastStatusAt pgtype.Timestamptz
	Status       pgtype.Text
	AgentPoolID  pgtype.Text
	CurrentJobs  int64
}

func (q *Queries) FindAgentsByOrganization(ctx context.Context, organizationName pgtype.Text) ([]FindAgentsByOrganizationRow, error) {
	rows, err := q.db.Query(ctx, findAgentsByOrganization, organizationName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []FindAgentsByOrganizationRow
	for rows.Next() {
		var i FindAgentsByOrganizationRow
		if err := rows.Scan(
			&i.AgentID,
			&i.Name,
			&i.Version,
			&i.MaxJobs,
			&i.IPAddress,
			&i.LastPingAt,
			&i.LastStatusAt,
			&i.Status,
			&i.AgentPoolID,
			&i.CurrentJobs,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const findAgentsByPoolID = `-- name: FindAgentsByPoolID :many
SELECT
    a.agent_id, a.name, a.version, a.max_jobs, a.ip_address, a.last_ping_at, a.last_status_at, a.status, a.agent_pool_id,
    ( SELECT count(*)
      FROM jobs j
      WHERE a.agent_id = j.agent_id
      AND j.status IN ('allocated', 'running')
    ) AS current_jobs
FROM agents a
JOIN agent_pools ap USING (agent_pool_id)
WHERE ap.agent_pool_id = $1
GROUP BY a.agent_id
ORDER BY last_ping_at DESC
`

type FindAgentsByPoolIDRow struct {
	AgentID      pgtype.Text
	Name         pgtype.Text
	Version      pgtype.Text
	MaxJobs      pgtype.Int4
	IPAddress    netip.Addr
	LastPingAt   pgtype.Timestamptz
	LastStatusAt pgtype.Timestamptz
	Status       pgtype.Text
	AgentPoolID  pgtype.Text
	CurrentJobs  int64
}

func (q *Queries) FindAgentsByPoolID(ctx context.Context, agentPoolID pgtype.Text) ([]FindAgentsByPoolIDRow, error) {
	rows, err := q.db.Query(ctx, findAgentsByPoolID, agentPoolID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []FindAgentsByPoolIDRow
	for rows.Next() {
		var i FindAgentsByPoolIDRow
		if err := rows.Scan(
			&i.AgentID,
			&i.Name,
			&i.Version,
			&i.MaxJobs,
			&i.IPAddress,
			&i.LastPingAt,
			&i.LastStatusAt,
			&i.Status,
			&i.AgentPoolID,
			&i.CurrentJobs,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const findServerAgents = `-- name: FindServerAgents :many
SELECT
    a.agent_id, a.name, a.version, a.max_jobs, a.ip_address, a.last_ping_at, a.last_status_at, a.status, a.agent_pool_id,
    ( SELECT count(*)
      FROM jobs j
      WHERE a.agent_id = j.agent_id
      AND j.status IN ('allocated', 'running')
    ) AS current_jobs
FROM agents a
WHERE agent_pool_id IS NULL
GROUP BY a.agent_id
ORDER BY last_ping_at DESC
`

type FindServerAgentsRow struct {
	AgentID      pgtype.Text
	Name         pgtype.Text
	Version      pgtype.Text
	MaxJobs      pgtype.Int4
	IPAddress    netip.Addr
	LastPingAt   pgtype.Timestamptz
	LastStatusAt pgtype.Timestamptz
	Status       pgtype.Text
	AgentPoolID  pgtype.Text
	CurrentJobs  int64
}

func (q *Queries) FindServerAgents(ctx context.Context) ([]FindServerAgentsRow, error) {
	rows, err := q.db.Query(ctx, findServerAgents)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []FindServerAgentsRow
	for rows.Next() {
		var i FindServerAgentsRow
		if err := rows.Scan(
			&i.AgentID,
			&i.Name,
			&i.Version,
			&i.MaxJobs,
			&i.IPAddress,
			&i.LastPingAt,
			&i.LastStatusAt,
			&i.Status,
			&i.AgentPoolID,
			&i.CurrentJobs,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const insertAgent = `-- name: InsertAgent :exec
INSERT INTO agents (
    agent_id,
    name,
    version,
    max_jobs,
    ip_address,
    last_ping_at,
    last_status_at,
    status,
    agent_pool_id
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8,
    $9
)
`

type InsertAgentParams struct {
	AgentID      pgtype.Text
	Name         pgtype.Text
	Version      pgtype.Text
	MaxJobs      pgtype.Int4
	IPAddress    netip.Addr
	LastPingAt   pgtype.Timestamptz
	LastStatusAt pgtype.Timestamptz
	Status       pgtype.Text
	AgentPoolID  pgtype.Text
}

func (q *Queries) InsertAgent(ctx context.Context, arg InsertAgentParams) error {
	_, err := q.db.Exec(ctx, insertAgent,
		arg.AgentID,
		arg.Name,
		arg.Version,
		arg.MaxJobs,
		arg.IPAddress,
		arg.LastPingAt,
		arg.LastStatusAt,
		arg.Status,
		arg.AgentPoolID,
	)
	return err
}

const updateAgent = `-- name: UpdateAgent :one
UPDATE agents
SET status = $1,
    last_ping_at = $2,
    last_status_at = $3
WHERE agent_id = $4
RETURNING agent_id, name, version, max_jobs, ip_address, last_ping_at, last_status_at, status, agent_pool_id
`

type UpdateAgentParams struct {
	Status       pgtype.Text
	LastPingAt   pgtype.Timestamptz
	LastStatusAt pgtype.Timestamptz
	AgentID      pgtype.Text
}

func (q *Queries) UpdateAgent(ctx context.Context, arg UpdateAgentParams) (Agent, error) {
	row := q.db.QueryRow(ctx, updateAgent,
		arg.Status,
		arg.LastPingAt,
		arg.LastStatusAt,
		arg.AgentID,
	)
	var i Agent
	err := row.Scan(
		&i.AgentID,
		&i.Name,
		&i.Version,
		&i.MaxJobs,
		&i.IPAddress,
		&i.LastPingAt,
		&i.LastStatusAt,
		&i.Status,
		&i.AgentPoolID,
	)
	return i, err
}
