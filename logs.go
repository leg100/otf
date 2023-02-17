package otf

import (
	"context"
	"fmt"
)

type LogService interface {
	// GetChunk retrieves a chunk of logs
	GetChunk(ctx context.Context, opts GetChunkOptions) (Chunk, error)
	// PutChunk persists a chunk of logs
	PutChunk(ctx context.Context, chunk Chunk) error
	// Tail follows a stream of log chunks
	Tail(ctx context.Context, opts GetChunkOptions) (<-chan Chunk, error)
}

type GetChunkOptions struct {
	RunID string    `schema:"run_id"`
	Phase PhaseType `schema:"phase"`
	// Limit is the size of the chunk to retrieve
	Limit int `schema:"limit"`
	// Offset is the position in the data from which to retrieve the chunk.
	Offset int `schema:"offset"`
}

// Key returns an identifier for looking up chunks in a cache
func (c GetChunkOptions) Key() string {
	return fmt.Sprintf("%s.%s.log", c.RunID, string(c.Phase))
}
