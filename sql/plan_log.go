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
func (db PlanLogDB) PutChunk(ctx context.Context, planID string, chunk otf.Chunk) error {
	q := NewQuerier(db.Conn)

	if len(chunk.Data) == 0 {
		return nil
	}

	data := chunk.Data

	if chunk.End {
		data = append(data, otf.ChunkEndMarker)
	}

	_, err := q.InsertPlanLogChunk(ctx, &planID, data)
	return err
}

// GetChunk retrieves a log chunk from the DB. GetChunkOptions is ignored, so
// the biggest chunk possible will be retrieved. Instead, it is expected an
// in-memory cache is placed in front, calling this method and caching the
// content.
func (db PlanLogDB) GetChunk(ctx context.Context, planID string, _ otf.GetChunkOptions) (otf.Chunk, error) {
	q := NewQuerier(db.Conn)

	data, err := q.FindPlanLogChunks(ctx, &planID)
	if err != nil {
		return otf.Chunk{}, err
	}

	chunk := otf.Chunk{Data: data}

	// NOTE: data should always be non-zero but empty data may have been
	// inserted into the DB in error.
	if len(data) > 0 {
		chunk.Start = true

		if data[len(data)-1] == otf.ChunkEndMarker {
			chunk.End = true
		}
	}

	return chunk, nil
}
