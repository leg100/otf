package otf

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"
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

// Stream retrieves chunks from the chunk store for an entity with id at regular
// intervals, and writes them to w, until the last chunk is retrieved.
func Stream(ctx context.Context, id string, store ChunkStore, w io.Writer, interval time.Duration, limit int) error {
	ckr := chunker{
		ChunkStore: store,
		w:          w,
		id:         id,
		limit:      limit,
	}

	for {
		select {
		case <-time.After(interval):
			last, err := ckr.write(ctx)
			if err != nil {
				return fmt.Errorf("writing chunk: %w", err)
			}
			if last {
				return nil
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// chunker writes chunks from the store to the wrapped writer
type chunker struct {
	ChunkStore
	offset, limit int
	w             io.Writer
	id            string
	last          bool
}

// writeChunk writes a chunk from the store; if the last chunk then it returns
// true.
func (c *chunker) write(ctx context.Context) (bool, error) {
	chunk, err := c.GetChunk(ctx, c.id, GetChunkOptions{Offset: c.offset, Limit: c.limit})
	if err != nil {
		return false, fmt.Errorf("retrieving chunk: %w", err)
	}

	if len(chunk) == 0 {
		return false, nil
	}

	c.offset += len(chunk)

	if bytes.HasPrefix(chunk, []byte{ChunkStartMarker}) {
		// Strip STX byte from chunk
		chunk = chunk[1:]
	}

	if bytes.HasSuffix(chunk, []byte{ChunkEndMarker}) {
		// Strip ETX byte from chunk
		chunk = chunk[:len(chunk)-1]
		c.last = true
	}

	_, err = c.w.Write(chunk)
	if err != nil {
		return false, fmt.Errorf("writing chunk: %w", err)
	}

	return c.last, nil
}
