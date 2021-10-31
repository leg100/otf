package sql

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/leg100/otf"
)

var (
	_ otf.ChunkStore = (*ApplyLogDB)(nil)
)

type ApplyLogDB struct {
	*sqlx.DB
}

func NewApplyLogDB(db *sqlx.DB) *ApplyLogDB {
	return &ApplyLogDB{
		DB: db,
	}
}

// Put persists a Log to the DB.
func (db ApplyLogDB) PutChunk(ctx context.Context, applyID string, chunk []byte, opts otf.PutChunkOptions) error {
	return putChunk(ctx, db, "apply_logs", "apply_id", applyID, chunk, opts)
}

func (db ApplyLogDB) GetChunk(ctx context.Context, applyID string, opts otf.GetChunkOptions) ([]byte, error) {
	return getChunk(ctx, db, "apply_logs", "apply_id", applyID, opts)
}
