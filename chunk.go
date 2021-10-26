package otf

import (
	"context"
)

// ChunkStore implementations provide a persistent store from and to which chunks
// of binary objects can be fetched and uploaded.
type ChunkStore interface {
	// GetChunk fetches a blob chunk
	GetChunk(ctx context.Context, id string, opts GetChunkOptions) ([]byte, error)

	// PutChunk uploads a blob chunk
	PutChunk(ctx context.Context, id string, chunk []byte, opts PutChunkOptions) error
}

type GetChunkOptions struct {
	// Limit is the size of the chunk to retrieve
	Limit int `schema:"limit"`

	// Offset is the position within the binary object to retrieve the chunk
	Offset int `schema:"offset"`
}

type PutChunkOptions struct {
	// Start indicates this is the first chunk
	Start bool `schema:"start"`

	// End indicates this is the last and final chunk
	End bool `schema:"end"`
}
