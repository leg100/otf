package sql

import (
	"context"
	"math"

	"github.com/leg100/otf"
	"github.com/leg100/otf/sql/pggen"
)

type applyLogStore struct {
	*DB
}

// PutChunk persists a plan log chunk to the DB.
func (db *applyLogStore) PutChunk(ctx context.Context, applyID string, chunk otf.Chunk) error {
	if len(chunk.Data) == 0 {
		return nil
	}
	_, err := db.InsertApplyLogChunk(ctx,
		String(applyID),
		chunk.Marshal(),
	)
	return err
}

// GetChunk retrieves a plan log chunk from the DB.
func (db *applyLogStore) GetChunk(ctx context.Context, applyID string, opts otf.GetChunkOptions) (otf.Chunk, error) {
	// 0 means limitless but in SQL it means 0 so as a workaround set it to the
	// maximum a postgres INT can hold.
	if opts.Limit == 0 {
		opts.Limit = math.MaxInt32
	}
	chunk, err := db.FindApplyLogChunks(ctx, pggen.FindApplyLogChunksParams{
		ApplyID: String(applyID),
		Offset:  opts.Offset + 1,
		Limit:   opts.Limit,
	})
	if err != nil {
		return otf.Chunk{}, err
	}
	return otf.UnmarshalChunk(chunk), nil
}

func (db *DB) ApplyLogStore() otf.ChunkStore {
	return &applyLogStore{DB: db}
}
