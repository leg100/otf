package logs

import (
	"context"
	"fmt"
	"math"

	"github.com/leg100/otf"
	"github.com/leg100/otf/sql"
	"github.com/leg100/otf/sql/pggen"
)

type db interface {
	// get fetches a chunk of logs.
	get(ctx context.Context, opts otf.GetChunkOptions) (otf.Chunk, error)
	// put uploads a chunk, receiving back the chunk along with a unique
	// ID.
	put(ctx context.Context, chunk otf.Chunk) (otf.PersistedChunk, error)
}

// pgdb is a logs database on postgres
type pgdb struct {
	otf.Database // provides access to generated SQL queries
}

func newPGDB(db otf.Database) *pgdb {
	return &pgdb{db}
}

// put persists a log chunk to the DB.
func (db *pgdb) put(ctx context.Context, chunk otf.Chunk) (otf.PersistedChunk, error) {
	if len(chunk.Data) == 0 {
		return otf.PersistedChunk{}, fmt.Errorf("refusing to persist empty chunk")
	}
	id, err := db.InsertLogChunk(ctx, pggen.InsertLogChunkParams{
		RunID:  sql.String(chunk.RunID),
		Phase:  sql.String(string(chunk.Phase)),
		Chunk:  chunk.Data,
		Offset: chunk.Offset,
	})
	if err != nil {
		return otf.PersistedChunk{}, sql.Error(err)
	}
	return otf.PersistedChunk{
		ChunkID: id,
		Chunk:   chunk,
	}, nil
}

// get retrieves a log chunk from the DB.
func (db *pgdb) get(ctx context.Context, opts otf.GetChunkOptions) (otf.Chunk, error) {
	// 0 means limitless but in SQL it means 0 so as a workaround set it to the
	// maximum a postgres INT can hold.
	if opts.Limit == 0 {
		opts.Limit = math.MaxInt32
	}
	data, err := db.FindLogChunks(ctx, pggen.FindLogChunksParams{
		RunID:  sql.String(opts.RunID),
		Phase:  sql.String(string(opts.Phase)),
		Offset: opts.Offset,
		Limit:  opts.Limit,
	})
	if err != nil {
		return otf.Chunk{}, sql.Error(err)
	}
	return otf.Chunk{
		RunID:  opts.RunID,
		Phase:  opts.Phase,
		Data:   data,
		Offset: opts.Offset,
	}, nil
}

// getByID retrieves a log chunk from the DB using its unique ID.
func (db *pgdb) getByID(ctx context.Context, chunkID int) (otf.PersistedChunk, error) {
	chunk, err := db.FindLogChunkByID(ctx, chunkID)
	if err != nil {
		return otf.PersistedChunk{}, sql.Error(err)
	}
	return otf.PersistedChunk{
		ChunkID: chunkID,
		Chunk: otf.Chunk{
			RunID:  chunk.RunID.String,
			Phase:  otf.PhaseType(chunk.Phase.String),
			Data:   chunk.Chunk,
			Offset: chunk.Offset,
		},
	}, nil
}
