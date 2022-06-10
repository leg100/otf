package sql

import (
	"context"
	"math"

	"github.com/jackc/pgtype"
	"github.com/leg100/otf"
	"github.com/leg100/otf/sql/pggen"
)

// PutChunk persists a log chunk to the DB.
func (db *DB) PutChunk(ctx context.Context, jobID string, chunk otf.Chunk) error {
	if len(chunk.Data) == 0 {
		return nil
	}
	_, err := db.InsertLogChunk(ctx,
		pgtype.Text{String: jobID, Status: pgtype.Present},
		chunk.Marshal(),
	)
	return err
}

// GetChunk retrieves a log chunk from the DB.
func (db *DB) GetChunk(ctx context.Context, jobID string, opts otf.GetChunkOptions) (otf.Chunk, error) {
	// 0 means limitless but in SQL it means 0 so as a workaround set it to the
	// maximum a postgres INT can hold.
	if opts.Limit == 0 {
		opts.Limit = math.MaxInt32
	}
	chunk, err := db.FindLogChunks(ctx, pggen.FindLogChunksParams{
		JobID:  pgtype.Text{String: jobID, Status: pgtype.Present},
		Offset: opts.Offset + 1,
		Limit:  opts.Limit,
	})
	if err != nil {
		return otf.Chunk{}, err
	}
	return otf.UnmarshalChunk(chunk), nil
}

// GetLogsByApplyID retrieves all log chunks for an apply from the DB.
func (db *DB) GetLogsByApplyID(ctx context.Context, applyID string) (otf.Chunk, error) {
	logs, err := db.FindAllLogChunksUsingApplyID(ctx, pgtype.Text{String: applyID, Status: pgtype.Present})
	if err != nil {
		return otf.Chunk{}, err
	}
	return otf.UnmarshalChunk(logs), nil
}
