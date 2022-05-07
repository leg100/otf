package sql

import (
	"context"

	"github.com/jackc/pgx/v4"
	"github.com/leg100/otf"
)

var (
	_ otf.ChunkStore = (*PlanLogDB)(nil)
)

type PlanLogDB struct {
	*pgx.Conn
}

func NewPlanLogDB(conn *pgx.Conn) *PlanLogDB {
	return &PlanLogDB{
		Conn: conn,
	}
}

// PutChunk persists a log chunk to the DB.
func (db PlanLogDB) PutChunk(ctx context.Context, planID string, chunk []byte, opts otf.PutChunkOptions) error {
	q := NewQuerier(db.Conn)

	_, err := q.InsertPlanLogChunk(ctx, InsertPlanLogChunkParams{
		PlanID: &planID,
		Chunk:  chunk,
		Start:  &opts.Start,
		End:    &opts.End,
		Size:   int32(len(chunk)),
	})
	return err
}

// GetChunk retrieves a log chunk from the DB.
func (db PlanLogDB) GetChunk(ctx context.Context, planID string, opts otf.GetChunkOptions) ([]byte, error) {
	q := NewQuerier(db.Conn)

	result, err := q.FindPlanLogChunks(ctx, &planID)
	if err != nil {
		return nil, err
	}

	return mergeChunks
}
