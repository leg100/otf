package logs

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/sql"
)

// pgdb is a logs database on postgres
type pgdb struct {
	*sql.DB
}

func (db *pgdb) putChunk(ctx context.Context, chunk Chunk) error {
	return db.Tx(ctx, func(ctx context.Context) error {
		_, err := db.Exec(ctx, `
INSERT INTO chunks (
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
		if err != nil {
			return err
		}
		if !chunk.IsEnd() {
			return nil
		}
		// Now that the last chunk of logs for a run phase has been inserted into the
		// chunks table, all the chunks for the run phase can be coalesced into
		// a single row in the *logs* table, and the chunks can be deleted.
		_, err = db.Exec(ctx, `
WITH chunks AS (
	SELECT
		string_agg(chunk, '') AS coalesced
	FROM (
		SELECT run_id, phase, chunk
		FROM chunks
		WHERE run_id = @run_id
		AND   phase  = @phase
		ORDER BY _offset
	)
	GROUP BY run_id, phase
)
INSERT INTO logs (run_id, phase, logs)
SELECT
    @run_id,
    @phase,
	coalesced
FROM chunks
`, pgx.NamedArgs{
			"run_id": chunk.RunID,
			"phase":  chunk.Phase,
		})
		if err != nil {
			return err
		}
		_, err = db.Exec(ctx, `
DELETE
FROM chunks
WHERE run_id = @run_id
AND   phase  = @phase
`, pgx.NamedArgs{
			"run_id": chunk.RunID,
			"phase":  chunk.Phase,
		})
		if err != nil {
			return err
		}
		return nil
	})
}

func (db *pgdb) getChunk(ctx context.Context, opts GetChunkOptions) (Chunk, error) {
	// Sanitize options. If the limit is 0 then interpret this to mean
	// limitless. In order to keep the SQL query below a hardcoded string we
	// interpret limitless to be the max field size in postgres, which reading
	// around appears to be 1GB. If there are logs for a run greater than that
	// then god help us.
	if opts.Limit == 0 {
		opts.Limit = 1024 ^ 4
	}
	rows := db.Query(ctx, `
SELECT
    substring(string_agg(chunk, '') from @offset + 1 for @limit)
FROM (
    SELECT run_id, phase, chunk
    FROM chunks
    WHERE run_id = @run_id
    AND   phase  = @phase
    ORDER BY _offset
)
GROUP BY run_id, phase
UNION
SELECT
    substring(logs from @offset + 1 for @limit)
FROM logs
WHERE run_id = @run_id
AND   phase  = @phase
`, pgx.NamedArgs{
		"run_id": opts.RunID,
		"phase":  opts.Phase,
		"limit":  opts.Limit,
		"offset": opts.Offset,
	})
	logs, err := sql.CollectOneRow(rows, pgx.RowTo[[]byte])
	if err != nil {
		// Don't consider no logs an error because logs may not have been
		// uploaded yet.
		if errors.Is(err, internal.ErrResourceNotFound) {
			chunk := Chunk{
				RunID:  opts.RunID,
				Phase:  opts.Phase,
				Offset: opts.Offset,
			}
			return chunk, nil
		}
		return Chunk{}, err
	}
	chunk := Chunk{
		RunID:  opts.RunID,
		Phase:  opts.Phase,
		Offset: opts.Offset,
		Data:   logs,
	}
	return chunk, nil
}
