package logs

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/sql"
)

// pgdb is a logs database on postgres
type pgdb struct {
	*sql.DB // provides access to generated SQL queries
}

func (db *pgdb) put(ctx context.Context, chunk Chunk) error {
	_, err := db.Exec(ctx, `
INSERT INTO logs (
    chunk_id,
    run_id,
    phase,
    chunk,
    _offset
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5
)`,
		chunk.ID,
		chunk.RunID,
		chunk.Phase,
		chunk.Data,
		chunk.Offset,
	)
	return err
}

func (db *pgdb) getLogs(ctx context.Context, runID resource.TfeID, phase internal.PhaseType) ([]byte, error) {
	rows := db.Query(ctx, `
SELECT
    string_agg(chunk, '')
FROM (
    SELECT run_id, phase, chunk
    FROM logs
    WHERE run_id = $1
    AND   phase  = $2
    ORDER BY _offset
) c
GROUP BY run_id, phase
`, runID, phase)
	logs, err := sql.CollectOneRow(rows, pgx.RowTo[[]byte])
	if err != nil {
		// Don't consider no logs an error because logs may not have been
		// uploaded yet.
		if errors.Is(err, internal.ErrResourceNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return logs, nil
}
