package otf

import (
	"context"
	"fmt"
)

// ChunkStore implementations provide a persistent store from and to which chunks
// of binary objects can be fetched and uploaded.
type ChunkStore interface {
	// GetChunk fetches a blob chunk for entity with id
	GetChunk(ctx context.Context, id string, opts GetChunkOptions) ([]byte, error)

	// PutChunk uploads a blob chunk for entity with id
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

// GetChunk returns a chunk of data.
func GetChunk(data []byte, opts GetChunkOptions) ([]byte, error) {
	if opts.Offset > len(data) {
		return nil, fmt.Errorf("chunk offset greater than size of data: %d > %d", opts.Offset, len(data))
	}

	if opts.Limit == 0 {
		return data[opts.Offset:], nil
	}

	if opts.Limit > ChunkMaxLimit {
		opts.Limit = ChunkMaxLimit
	}

	// Adjust limit if it extends beyond size of value
	if (opts.Offset + opts.Limit) > len(data) {
		opts.Limit = len(data) - opts.Offset
	}

	return data[opts.Offset:(opts.Offset + opts.Limit)], nil
}
