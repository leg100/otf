package logs

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/sql"
)

var q = &Queries{}

// pgdb is a logs database on postgres
type pgdb struct {
	*sql.DB // provides access to generated SQL queries
}

func (db *pgdb) put(ctx context.Context, chunk Chunk) error {
	err := q.InsertLogChunk(ctx, db.Conn(ctx), InsertLogChunkParams{
		ChunkID: chunk.TfeID,
		RunID:   chunk.RunID,
		Phase:   sql.String(string(chunk.Phase)),
		Chunk:   chunk.Data,
		Offset:  sql.Int4(chunk.Offset),
	})
	if err != nil {
		return sql.Error(err)
	}
	return nil
}

func (db *pgdb) getChunk(ctx context.Context, chunkID resource.TfeID) (Chunk, error) {
	chunk, err := q.FindLogChunkByID(ctx, db.Conn(ctx), chunkID)
	if err != nil {
		return Chunk{}, sql.Error(err)
	}
	return Chunk{
		TfeID:  chunkID,
		RunID:  chunk.RunID,
		Phase:  internal.PhaseType(chunk.Phase.String),
		Data:   chunk.Chunk,
		Offset: int(chunk.Offset.Int32),
	}, nil
}

func (db *pgdb) getLogs(ctx context.Context, runID resource.TfeID, phase internal.PhaseType) ([]byte, error) {
	data, err := q.FindLogs(ctx, db.Conn(ctx), FindLogsParams{
		RunID: runID,
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
