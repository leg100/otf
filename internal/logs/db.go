package logs

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/jackc/pgtype"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/sql/pggen"
)

type (

	// pgdb is a logs database on postgres
	pgdb struct {
		internal.DB // provides access to generated SQL queries
	}

	pgresult struct {
		ChunkID int         `json:"chunk_id"`
		RunID   pgtype.Text `json:"run_id"`
		Phase   pgtype.Text `json:"phase"`
		Chunk   []byte      `json:"chunk"`
		Offset  int         `json:"offset"`
	}
)

// UnmarshalEvent implements EventUnmarshaler
func (db *pgdb) UnmarshalEvent(ctx context.Context, payload []byte, op internal.EventType) (any, error) {
	var r pgresult
	err := json.Unmarshal(payload, &r)
	if err != nil {
		return nil, err
	}
	return internal.Chunk{
		ID:     strconv.Itoa(r.ChunkID),
		RunID:  r.RunID.String,
		Phase:  internal.PhaseType(r.Phase.String),
		Data:   r.Chunk,
		Offset: r.Offset,
	}, nil
}

// put persists data to the DB and returns a unique identifier for the chunk
func (db *pgdb) put(ctx context.Context, opts internal.PutChunkOptions) (string, error) {
	if len(opts.Data) == 0 {
		return "", fmt.Errorf("refusing to persist empty chunk")
	}
	id, err := db.InsertLogChunk(ctx, pggen.InsertLogChunkParams{
		RunID:  sql.String(opts.RunID),
		Phase:  sql.String(string(opts.Phase)),
		Chunk:  opts.Data,
		Offset: opts.Offset,
	})
	if err != nil {
		return "", sql.Error(err)
	}
	return strconv.Itoa(id), nil
}
