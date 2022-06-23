package sql

import (
	"context"
	"math"

	"github.com/leg100/otf"
	"github.com/leg100/otf/sql/pggen"
)

type planLogStore struct {
	*DB
}

// PutChunk persists a plan log chunk to the DB.
func (db *planLogStore) PutChunk(ctx context.Context, planID string, chunk otf.Chunk) error {
	if len(chunk.Data) == 0 {
		return nil
	}
	_, err := db.InsertPlanLogChunk(ctx, String(planID), chunk.Marshal())
	return err
}

// GetChunk retrieves a plan log chunk from the DB.
func (db *planLogStore) GetChunk(ctx context.Context, planID string, opts otf.GetChunkOptions) (otf.Chunk, error) {
	// 0 means limitless but in SQL it means 0 so as a workaround set it to the
	// maximum a postgres INT can hold.
	if opts.Limit == 0 {
		opts.Limit = math.MaxInt32
	}
	chunk, err := db.FindPlanLogChunks(ctx, pggen.FindPlanLogChunksParams{
		PlanID: String(planID),
		Offset: opts.Offset + 1,
		Limit:  opts.Limit,
	})
	if err != nil {
		return otf.Chunk{}, err
	}
	return otf.UnmarshalChunk(chunk), nil
}

func (db *DB) PlanLogStore() otf.ChunkStore {
	return &planLogStore{DB: db}
}
