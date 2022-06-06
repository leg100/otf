package sql

import (
	"context"
	"math"

	"github.com/jackc/pgtype"
	"github.com/leg100/otf"
	"github.com/leg100/otf/sql/pggen"
)

type ApplyLogsDB struct {
	*DB
}

func (db *DB) ApplyLogs() otf.ChunkStore {
	return &ApplyLogsDB{DB: db}
}

// PutChunk persists a log chunk to the DB.
func (db ApplyLogsDB) PutChunk(ctx context.Context, applyID string, chunk otf.Chunk) error {
	if len(chunk.Data) == 0 {
		return nil
	}
	_, err := db.InsertApplyLogChunk(ctx, pgtype.Text{String: applyID, Status: pgtype.Present}, chunk.Marshal())
	return err
}

// GetChunk retrieves a log chunk from the DB.
func (db ApplyLogsDB) GetChunk(ctx context.Context, applyID string, opts otf.GetChunkOptions) (otf.Chunk, error) {
	// 0 means limitless but in SQL it means 0 so as a workaround set it to the
	// maximum a postgres INT can hold.
	if opts.Limit == 0 {
		opts.Limit = math.MaxInt32
	}
	chunk, err := db.FindApplyLogChunks(ctx, pggen.FindApplyLogChunksParams{
		ApplyID: pgtype.Text{String: applyID, Status: pgtype.Present},
		Offset:  opts.Offset + 1,
		Limit:   opts.Limit,
	})
	if err != nil {
		return otf.Chunk{}, err
	}

	return otf.UnmarshalChunk(chunk), nil
}
