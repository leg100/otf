// Code generated by pggen. DO NOT EDIT.

package sql

import (
	"context"
	"fmt"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"time"
)

const insertTokenSQL = `INSERT INTO tokens (
    token_id,
    token,
    created_at,
    updated_at,
    description,
    user_id
) VALUES (
    $1,
    $2,
    current_timestamp,
    current_timestamp,
    $3,
    $4
)
RETURNING *;`

type InsertTokenParams struct {
	TokenID     string
	Token       string
	Description string
	UserID      string
}

type InsertTokenRow struct {
	TokenID     string    `json:"token_id"`
	Token       *string   `json:"token"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Description string    `json:"description"`
	UserID      string    `json:"user_id"`
}

func (s InsertTokenRow) GetTokenID() string { return s.TokenID }
func (s InsertTokenRow) GetToken() *string { return s.Token }
func (s InsertTokenRow) GetCreatedAt() time.Time { return s.CreatedAt }
func (s InsertTokenRow) GetUpdatedAt() time.Time { return s.UpdatedAt }
func (s InsertTokenRow) GetDescription() string { return s.Description }
func (s InsertTokenRow) GetUserID() string { return s.UserID }


// InsertToken implements Querier.InsertToken.
func (q *DBQuerier) InsertToken(ctx context.Context, params InsertTokenParams) (InsertTokenRow, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "InsertToken")
	row := q.conn.QueryRow(ctx, insertTokenSQL, params.TokenID, params.Token, params.Description, params.UserID)
	var item InsertTokenRow
	if err := row.Scan(&item.TokenID, &item.Token, &item.CreatedAt, &item.UpdatedAt, &item.Description, &item.UserID); err != nil {
		return item, fmt.Errorf("query InsertToken: %w", err)
	}
	return item, nil
}

// InsertTokenBatch implements Querier.InsertTokenBatch.
func (q *DBQuerier) InsertTokenBatch(batch genericBatch, params InsertTokenParams) {
	batch.Queue(insertTokenSQL, params.TokenID, params.Token, params.Description, params.UserID)
}

// InsertTokenScan implements Querier.InsertTokenScan.
func (q *DBQuerier) InsertTokenScan(results pgx.BatchResults) (InsertTokenRow, error) {
	row := results.QueryRow()
	var item InsertTokenRow
	if err := row.Scan(&item.TokenID, &item.Token, &item.CreatedAt, &item.UpdatedAt, &item.Description, &item.UserID); err != nil {
		return item, fmt.Errorf("scan InsertTokenBatch row: %w", err)
	}
	return item, nil
}

const deleteTokenByIDSQL = `DELETE
FROM tokens
WHERE token_id = $1;`

// DeleteTokenByID implements Querier.DeleteTokenByID.
func (q *DBQuerier) DeleteTokenByID(ctx context.Context, tokenID string) (pgconn.CommandTag, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "DeleteTokenByID")
	cmdTag, err := q.conn.Exec(ctx, deleteTokenByIDSQL, tokenID)
	if err != nil {
		return cmdTag, fmt.Errorf("exec query DeleteTokenByID: %w", err)
	}
	return cmdTag, err
}

// DeleteTokenByIDBatch implements Querier.DeleteTokenByIDBatch.
func (q *DBQuerier) DeleteTokenByIDBatch(batch genericBatch, tokenID string) {
	batch.Queue(deleteTokenByIDSQL, tokenID)
}

// DeleteTokenByIDScan implements Querier.DeleteTokenByIDScan.
func (q *DBQuerier) DeleteTokenByIDScan(results pgx.BatchResults) (pgconn.CommandTag, error) {
	cmdTag, err := results.Exec()
	if err != nil {
		return cmdTag, fmt.Errorf("exec DeleteTokenByIDBatch: %w", err)
	}
	return cmdTag, err
}
