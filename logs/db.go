package sql

import (
	"context"
	"fmt"
	"math"

	"github.com/leg100/otf"
	"github.com/leg100/otf/sql/pggen"
)

// PutChunk persists a log chunk to the DB.
func (db *DB) PutChunk(ctx context.Context, chunk otf.Chunk) (otf.PersistedChunk, error) {
	if len(chunk.Data) == 0 {
		return otf.PersistedChunk{}, fmt.Errorf("refusing to persist empty chunk")
	}
	id, err := db.InsertLogChunk(ctx, pggen.InsertLogChunkParams{
		RunID:  String(chunk.RunID),
		Phase:  String(string(chunk.Phase)),
		Chunk:  chunk.Data,
		Offset: chunk.Offset,
	})
	if err != nil {
		return otf.PersistedChunk{}, Error(err)
	}
	return otf.PersistedChunk{
		ChunkID: id,
		Chunk:   chunk,
	}, nil
}

// GetChunk retrieves a log chunk from the DB.
func (db *DB) GetChunk(ctx context.Context, opts otf.GetChunkOptions) (otf.Chunk, error) {
	// 0 means limitless but in SQL it means 0 so as a workaround set it to the
	// maximum a postgres INT can hold.
	if opts.Limit == 0 {
		opts.Limit = math.MaxInt32
	}
	data, err := db.FindLogChunks(ctx, pggen.FindLogChunksParams{
		RunID:  String(opts.RunID),
		Phase:  String(string(opts.Phase)),
		Offset: opts.Offset,
		Limit:  opts.Limit,
	})
	if err != nil {
		return otf.Chunk{}, Error(err)
	}
	return otf.Chunk{
		RunID:  opts.RunID,
		Phase:  opts.Phase,
		Data:   data,
		Offset: opts.Offset,
	}, nil
}

// GetChunkByID retrieves a plan log chunk from the DB using its unique chunk ID.
func (db *DB) GetChunkByID(ctx context.Context, chunkID int) (otf.PersistedChunk, error) {
	chunk, err := db.FindLogChunkByID(ctx, chunkID)
	if err != nil {
		return otf.PersistedChunk{}, Error(err)
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
