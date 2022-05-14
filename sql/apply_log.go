package sql

import (
	"context"
	"math"

	"github.com/jackc/pgx/v4"
	"github.com/leg100/otf"
)

var (
	_ otf.ChunkStore = (*ApplyLogDB)(nil)
)

type ApplyLogDB struct {
	*pgx.Conn
}

func NewApplyLogDB(conn *pgx.Conn) *ApplyLogDB {
	return &ApplyLogDB{
		Conn: conn,
	}
}

// PutChunk persists a log chunk to the DB.
func (db ApplyLogDB) PutChunk(ctx context.Context, applyID string, chunk otf.Chunk) error {
	q := NewQuerier(db.Conn)

	if len(chunk.Data) == 0 {
		return nil
	}

	_, err := q.InsertApplyLogChunk(ctx, applyID, chunk.Marshal())
	return err
}

// GetChunk retrieves a log chunk from the DB.
func (db ApplyLogDB) GetChunk(ctx context.Context, applyID string, opts otf.GetChunkOptions) (otf.Chunk, error) {
	q := NewQuerier(db.Conn)

	// 0 means limitless but in SQL it means 0 so as a workaround set it to the
	// maximum a postgres INT can hold.
	if opts.Limit == 0 {
		opts.Limit = math.MaxInt32
	}

	chunk, err := q.FindApplyLogChunks(ctx, FindApplyLogChunksParams{
		ApplyID: applyID,
		Offset:  int32(opts.Offset) + 1,
		Limit:   int32(opts.Limit),
	})
	if err != nil {
		return otf.Chunk{}, err
	}

	return otf.UnmarshalChunk(chunk), nil
}
