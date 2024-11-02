package logs

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/sql/sqlc"
)

// pgdb is a logs database on postgres
type pgdb struct {
	*sql.DB // provides access to generated SQL queries
}

// put persists a chunk of logs to the DB and returns the chunk updated with a
// unique identifier

// put persists data to the DB and returns a unique identifier for the chunk
func (db *pgdb) put(ctx context.Context, opts internal.PutChunkOptions) (string, error) {
	if len(opts.Data) == 0 {
		return "", fmt.Errorf("refusing to persist empty chunk")
	}
	id, err := db.Querier(ctx).InsertLogChunk(ctx, sqlc.InsertLogChunkParams{
		RunID:  sql.String(opts.RunID.String()),
		Phase:  sql.String(string(opts.Phase)),
		Chunk:  opts.Data,
		Offset: sql.Int4(opts.Offset),
	})
	if err != nil {
		return "", sql.Error(err)
	}
	return strconv.Itoa(int(id.Int32)), nil
}

func (db *pgdb) getChunk(ctx context.Context, chunkID resource.ID) (internal.Chunk, error) {
	id, err := strconv.Atoi(chunkID)
	if err != nil {
		return internal.Chunk{}, err
	}
	chunk, err := db.Querier(ctx).FindLogChunkByID(ctx, sql.Int4(id))
	if err != nil {
		return internal.Chunk{}, sql.Error(err)
	}
	return internal.Chunk{
		ID:     chunkID,
		RunID:  chunk.RunID.String,
		Phase:  internal.PhaseType(chunk.Phase.String),
		Data:   chunk.Chunk,
		Offset: int(chunk.Offset.Int32),
	}, nil
}

func (db *pgdb) getLogs(ctx context.Context, runID resource.ID, phase internal.PhaseType) ([]byte, error) {
	data, err := db.Querier(ctx).FindLogs(ctx, sqlc.FindLogsParams{
		RunID: sql.String(runID.String()),
		Phase: sql.String(string(phase)),
	})
	if err != nil {
		// Don't consider no rows an error because logs may not have been
		// uploaded yet.
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, sql.Error(err)
	}
	return data, nil
}
