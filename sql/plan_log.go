package sql

import (
	"context"
	"math"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/leg100/otf"
)

var (
	_ otf.ChunkStore = (*PlanLogDB)(nil)
)

type PlanLogDB struct {
	*pgxpool.Pool
}

func NewPlanLogDB(conn *pgxpool.Pool) *PlanLogDB {
	return &PlanLogDB{
		Pool: conn,
	}
}

// PutChunk persists a log chunk to the DB.
func (db PlanLogDB) PutChunk(ctx context.Context, planID string, chunk otf.Chunk) error {
	q := NewQuerier(db.Pool)

	if len(chunk.Data) == 0 {
		return nil
	}

	_, err := q.InsertPlanLogChunk(ctx, planID, chunk.Marshal())
	return err
}

// GetChunk retrieves a log chunk from the DB.
func (db PlanLogDB) GetChunk(ctx context.Context, planID string, opts otf.GetChunkOptions) (otf.Chunk, error) {
	q := NewQuerier(db.Pool)

	// 0 means limitless but in SQL it means 0 so as a workaround set it to the
	// maximum a postgres INT can hold.
	if opts.Limit == 0 {
		opts.Limit = math.MaxInt32
	}

	chunk, err := q.FindPlanLogChunks(ctx, FindPlanLogChunksParams{
		PlanID: planID,
		Offset: int32(opts.Offset) + 1,
		Limit:  int32(opts.Limit),
	})
	if err != nil {
		return otf.Chunk{}, err
	}

	return otf.UnmarshalChunk(chunk), nil
}
