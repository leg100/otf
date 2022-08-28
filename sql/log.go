package sql

import (
	"context"
	"math"

	"github.com/leg100/otf"
	"github.com/leg100/otf/sql/pggen"
)

// PutChunk persists a plan log chunk to the DB.
func (db *DB) PutChunk(ctx context.Context, runID string, phase otf.PhaseType, chunk otf.Chunk) error {
	if !chunk.Start && !chunk.End && len(chunk.Data) == 0 {
		// skip empty, marker-less chunks
		return nil
	}
	_, err := db.InsertLogChunk(ctx, pggen.InsertLogChunkParams{
		RunID: String(runID),
		Phase: String(string(phase)),
		Chunk: chunk.Marshal(),
	})
	return err
}

// GetChunk retrieves a plan log chunk from the DB.
func (db *DB) GetChunk(ctx context.Context, runID string, phase otf.PhaseType, opts otf.GetChunkOptions) (otf.Chunk, error) {
	// 0 means limitless but in SQL it means 0 so as a workaround set it to the
	// maximum a postgres INT can hold.
	if opts.Limit == 0 {
		opts.Limit = math.MaxInt32
	}
	chunk, err := db.FindLogChunks(ctx, pggen.FindLogChunksParams{
		RunID: String(runID),
		Phase: String(string(phase)),
		// TODO: we add one here because postgres' substr() starts from 1 not 0
		// - we should probably move this to the SQL itself
		Offset: opts.Offset + 1,
		Limit:  opts.Limit,
	})
	if err != nil {
		return otf.Chunk{}, databaseError(err)
	}
	return otf.UnmarshalChunk(chunk), nil
}
