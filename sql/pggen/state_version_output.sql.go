// Code generated by pggen. DO NOT EDIT.

package pggen

import (
	"context"
	"fmt"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
)

const insertStateVersionOutputSQL = `INSERT INTO state_version_outputs (
    state_version_output_id,
    name,
    sensitive,
    type,
    value,
    state_version_id
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6
);`

type InsertStateVersionOutputParams struct {
	ID             pgtype.Text
	Name           pgtype.Text
	Sensitive      bool
	Type           pgtype.Text
	Value          pgtype.Text
	StateVersionID pgtype.Text
}

// InsertStateVersionOutput implements Querier.InsertStateVersionOutput.
func (q *DBQuerier) InsertStateVersionOutput(ctx context.Context, params InsertStateVersionOutputParams) (pgconn.CommandTag, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "InsertStateVersionOutput")
	cmdTag, err := q.conn.Exec(ctx, insertStateVersionOutputSQL, params.ID, params.Name, params.Sensitive, params.Type, params.Value, params.StateVersionID)
	if err != nil {
		return cmdTag, fmt.Errorf("exec query InsertStateVersionOutput: %w", err)
	}
	return cmdTag, err
}

// InsertStateVersionOutputBatch implements Querier.InsertStateVersionOutputBatch.
func (q *DBQuerier) InsertStateVersionOutputBatch(batch genericBatch, params InsertStateVersionOutputParams) {
	batch.Queue(insertStateVersionOutputSQL, params.ID, params.Name, params.Sensitive, params.Type, params.Value, params.StateVersionID)
}

// InsertStateVersionOutputScan implements Querier.InsertStateVersionOutputScan.
func (q *DBQuerier) InsertStateVersionOutputScan(results pgx.BatchResults) (pgconn.CommandTag, error) {
	cmdTag, err := results.Exec()
	if err != nil {
		return cmdTag, fmt.Errorf("exec InsertStateVersionOutputBatch: %w", err)
	}
	return cmdTag, err
}
