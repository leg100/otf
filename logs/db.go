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
	// GetChunk fetches a chunk of logs.
	GetChunk(ctx context.Context, opts GetChunkOptions) (Chunk, error)
	// GetChunkByID fetches a specific chunk with the given ID.
	GetChunkByID(ctx context.Context, id int) (PersistedChunk, error)
	// PutChunk uploads a chunk, receiving back the chunk along with a unique
	// ID.
	PutChunk(ctx context.Context, chunk Chunk) (PersistedChunk, error)
}

// pgdb is a state/state-version database on postgres
type pgdb struct {
	otf.Database // provides access to generated SQL queries
}

func newPGDB(db otf.Database) *pgdb {
	return &pgdb{db}
}

// PutChunk persists a log chunk to the DB.
func (db *pgdb) PutChunk(ctx context.Context, chunk Chunk) (PersistedChunk, error) {
	if len(chunk.Data) == 0 {
		return PersistedChunk{}, fmt.Errorf("refusing to persist empty chunk")
	}
	id, err := db.InsertLogChunk(ctx, pggen.InsertLogChunkParams{
		RunID:  sql.String(chunk.RunID),
		Phase:  sql.String(string(chunk.Phase)),
		Chunk:  chunk.Data,
		Offset: chunk.Offset,
	})
	if err != nil {
		return PersistedChunk{}, sql.Error(err)
	}
	return PersistedChunk{
		ChunkID: id,
		Chunk:   chunk,
	}, nil
}

// GetChunk retrieves a log chunk from the DB.
func (db *pgdb) GetChunk(ctx context.Context, opts GetChunkOptions) (Chunk, error) {
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
		return Chunk{}, sql.Error(err)
	}
	return Chunk{
		RunID:  opts.RunID,
		Phase:  opts.Phase,
		Data:   data,
		Offset: opts.Offset,
	}, nil
}

// GetChunkByID retrieves a plan log chunk from the DB using its unique chunk ID.
func (db *pgdb) GetChunkByID(ctx context.Context, chunkID int) (PersistedChunk, error) {
	chunk, err := db.FindLogChunkByID(ctx, chunkID)
	if err != nil {
		return PersistedChunk{}, sql.Error(err)
	}
	return PersistedChunk{
		ChunkID: chunkID,
		Chunk: Chunk{
			RunID:  chunk.RunID.String,
			Phase:  otf.PhaseType(chunk.Phase.String),
			Data:   chunk.Chunk,
			Offset: chunk.Offset,
		},
	}, nil
}
