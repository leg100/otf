// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: agent_token.sql

package sqlc

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/leg100/otf/internal/resource"
)

const deleteAgentTokenByID = `-- name: DeleteAgentTokenByID :one
DELETE
FROM agent_tokens
WHERE agent_token_id = $1
RETURNING agent_token_id
`

func (q *Queries) DeleteAgentTokenByID(ctx context.Context, agentTokenID resource.ID) (resource.ID, error) {
	row := q.db.QueryRow(ctx, deleteAgentTokenByID, agentTokenID)
	var agent_token_id resource.ID
	err := row.Scan(&agent_token_id)
	return agent_token_id, err
}

const findAgentTokenByID = `-- name: FindAgentTokenByID :one
SELECT agent_token_id, created_at, description, agent_pool_id
FROM agent_tokens
WHERE agent_token_id = $1
`

func (q *Queries) FindAgentTokenByID(ctx context.Context, agentTokenID resource.ID) (AgentToken, error) {
	row := q.db.QueryRow(ctx, findAgentTokenByID, agentTokenID)
	var i AgentToken
	err := row.Scan(
		&i.AgentTokenID,
		&i.CreatedAt,
		&i.Description,
		&i.AgentPoolID,
	)
	return i, err
}

const findAgentTokensByAgentPoolID = `-- name: FindAgentTokensByAgentPoolID :many
SELECT agent_token_id, created_at, description, agent_pool_id
FROM agent_tokens
WHERE agent_pool_id = $1
ORDER BY created_at DESC
`

func (q *Queries) FindAgentTokensByAgentPoolID(ctx context.Context, agentPoolID resource.ID) ([]AgentToken, error) {
	rows, err := q.db.Query(ctx, findAgentTokensByAgentPoolID, agentPoolID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []AgentToken
	for rows.Next() {
		var i AgentToken
		if err := rows.Scan(
			&i.AgentTokenID,
			&i.CreatedAt,
			&i.Description,
			&i.AgentPoolID,
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

const insertAgentToken = `-- name: InsertAgentToken :exec
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
)
`

type InsertAgentTokenParams struct {
	AgentTokenID resource.ID
	CreatedAt    pgtype.Timestamptz
	Description  pgtype.Text
	AgentPoolID  resource.ID
}

func (q *Queries) InsertAgentToken(ctx context.Context, arg InsertAgentTokenParams) error {
	_, err := q.db.Exec(ctx, insertAgentToken,
		arg.AgentTokenID,
		arg.CreatedAt,
		arg.Description,
		arg.AgentPoolID,
	)
	return err
}
