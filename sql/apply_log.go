package sql

import (
	"context"
	"math"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/leg100/otf"
	"github.com/leg100/otf/sql/pggen"
)

var (
	_ otf.ChunkStore = (*ApplyLogDB)(nil)
)

type ApplyLogDB struct {
	*pgxpool.Pool
}

func NewApplyLogDB(conn *pgxpool.Pool) *ApplyLogDB {
	return &ApplyLogDB{
		Pool: conn,
	}
}

// PutChunk persists a log chunk to the DB.
func (db ApplyLogDB) PutChunk(ctx context.Context, applyID string, chunk otf.Chunk) error {
	q := pggen.NewQuerier(db.Pool)

	if len(chunk.Data) == 0 {
		return nil
	}

	_, err := q.InsertApplyLogChunk(ctx, pgtype.Text{String: applyID, Status: pgtype.Present}, chunk.Marshal())
	return err
}

// GetChunk retrieves a log chunk from the DB.
func (db ApplyLogDB) GetChunk(ctx context.Context, applyID string, opts otf.GetChunkOptions) (otf.Chunk, error) {
	q := pggen.NewQuerier(db.Pool)

	// 0 means limitless but in SQL it means 0 so as a workaround set it to the
	// maximum a postgres INT can hold.
	if opts.Limit == 0 {
		opts.Limit = math.MaxInt32
	}

	chunk, err := q.FindApplyLogChunks(ctx, pggen.FindApplyLogChunksParams{
		ApplyID: pgtype.Text{String: applyID, Status: pgtype.Present},
		Offset:  opts.Offset + 1,
		Limit:   opts.Limit,
	})
	if err != nil {
		return otf.Chunk{}, err
	}

	return otf.UnmarshalChunk(chunk), nil
}
