package sql

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/leg100/otf"
)

var (
	_ otf.ChunkStore = (*PlanLogDB)(nil)
)

type PlanLogDB struct {
	*sqlx.DB
}

func NewPlanLogDB(db *sqlx.DB) *PlanLogDB {
	return &PlanLogDB{
		DB: db,
	}
}

// PutChunk persists a log chunk to the DB.
func (db PlanLogDB) PutChunk(ctx context.Context, planID string, chunk []byte, opts otf.PutChunkOptions) error {
	return putChunk(ctx, db, "plan_logs", "plan_id", planID, chunk, opts)
}

// GetChunk retrieves a log chunk from the DB.
func (db PlanLogDB) GetChunk(ctx context.Context, planID string, opts otf.GetChunkOptions) ([]byte, error) {
	return getChunk(ctx, db, "plan_logs", "plan_id", planID, opts)
}
