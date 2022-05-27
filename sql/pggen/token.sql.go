// Code generated by pggen. DO NOT EDIT.

package pggen

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
)

const insertTokenSQL = `INSERT INTO tokens (
    token_id,
    token,
    created_at,
    description,
    user_id
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5
);`

type InsertTokenParams struct {
	TokenID     pgtype.Text
	Token       pgtype.Text
	CreatedAt   time.Time
	Description pgtype.Text
	UserID      pgtype.Text
}

// InsertToken implements Querier.InsertToken.
func (q *DBQuerier) InsertToken(ctx context.Context, params InsertTokenParams) (pgconn.CommandTag, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "InsertToken")
	cmdTag, err := q.conn.Exec(ctx, insertTokenSQL, params.TokenID, params.Token, params.CreatedAt, params.Description, params.UserID)
	if err != nil {
		return cmdTag, fmt.Errorf("exec query InsertToken: %w", err)
	}
	return cmdTag, err
}

// InsertTokenBatch implements Querier.InsertTokenBatch.
func (q *DBQuerier) InsertTokenBatch(batch genericBatch, params InsertTokenParams) {
	batch.Queue(insertTokenSQL, params.TokenID, params.Token, params.CreatedAt, params.Description, params.UserID)
}

// InsertTokenScan implements Querier.InsertTokenScan.
func (q *DBQuerier) InsertTokenScan(results pgx.BatchResults) (pgconn.CommandTag, error) {
	cmdTag, err := results.Exec()
	if err != nil {
		return cmdTag, fmt.Errorf("exec InsertTokenBatch: %w", err)
	}
	return cmdTag, err
}

const deleteTokenByIDSQL = `DELETE
FROM tokens
WHERE token_id = $1;`

// DeleteTokenByID implements Querier.DeleteTokenByID.
func (q *DBQuerier) DeleteTokenByID(ctx context.Context, tokenID pgtype.Text) (pgconn.CommandTag, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "DeleteTokenByID")
	cmdTag, err := q.conn.Exec(ctx, deleteTokenByIDSQL, tokenID)
	if err != nil {
		return cmdTag, fmt.Errorf("exec query DeleteTokenByID: %w", err)
	}
	return cmdTag, err
}

// DeleteTokenByIDBatch implements Querier.DeleteTokenByIDBatch.
func (q *DBQuerier) DeleteTokenByIDBatch(batch genericBatch, tokenID pgtype.Text) {
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
