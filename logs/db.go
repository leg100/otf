package logs

import (
	"context"
	"fmt"
	"strconv"

	"github.com/leg100/otf"
	"github.com/leg100/otf/sql"
	"github.com/leg100/otf/sql/pggen"
)

// pgdb is a logs database on postgres
type pgdb struct {
	otf.DB // provides access to generated SQL queries
}

// put persists a chunk of logs to the DB and returns the chunk updated with a
// unique identifier
func (db *pgdb) put(ctx context.Context, chunk otf.Chunk) (otf.Chunk, error) {
	if len(chunk.Data) == 0 {
		return otf.Chunk{}, fmt.Errorf("refusing to persist empty chunk")
	}
	id, err := db.InsertLogChunk(ctx, pggen.InsertLogChunkParams{
		RunID:  sql.String(chunk.RunID),
		Phase:  sql.String(string(chunk.Phase)),
		Chunk:  chunk.Data,
		Offset: chunk.Offset,
	})
	if err != nil {
		return otf.Chunk{}, sql.Error(err)
	}
	chunk.ID = strconv.Itoa(id)
	return chunk, nil
}

// GetByID implements pubsub.Getter
func (db *pgdb) GetByID(ctx context.Context, chunkID string) (any, error) {
	id, err := strconv.Atoi(chunkID)
	if err != nil {
		return nil, err
	}
	chunk, err := db.FindLogChunkByID(ctx, id)
	if err != nil {
		return nil, sql.Error(err)
	}
	return otf.Chunk{
		ID:     chunkID,
		RunID:  chunk.RunID.String,
		Phase:  otf.PhaseType(chunk.Phase.String),
		Data:   chunk.Chunk,
		Offset: chunk.Offset,
	}, nil
}
