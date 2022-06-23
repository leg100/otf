package otf

import (
	"context"
	"fmt"
)

// Chunk is a continuous portion of binary data, with start and end indicating
// if the portion includes the start and/or end of the binary data.
type Chunk struct {
	Data       []byte
	Start, End bool
}

// ChunkService has same signatures as ChunkStore
type ChunkService ChunkStore

// LogService is an alias for ChunkService
type LogService ChunkService

// ChunkStore implementations provide a persistent store from and to which chunks
// of binary objects can be fetched and uploaded.
type ChunkStore interface {
	// GetChunk fetches a blob chunk for entity with id
	GetChunk(ctx context.Context, id string, opts GetChunkOptions) (Chunk, error)
	// PutChunk uploads a blob chunk for entity with id
	PutChunk(ctx context.Context, id string, chunk Chunk) error
}

type GetChunkOptions struct {
	// Limit is the size of the chunk to retrieve
	Limit int `schema:"limit"`

	// Offset is the position within the binary object to retrieve the chunk.
	// NOTE: this includes the start and end marker bytes in an marshalled
	// chunk.
	Offset int `schema:"offset"`
}

type PutChunkOptions struct {
	// Start indicates this is the first chunk
	Start bool `schema:"start"`

	// End indicates this is the last and final chunk
	End bool `schema:"end"`
}

func (c Chunk) Marshal() []byte {
	if c.Start {
		c.Data = append([]byte{ChunkStartMarker}, c.Data...)
	}
	if c.End {
		c.Data = append(c.Data, ChunkEndMarker)
	}
	return c.Data
}

func UnmarshalChunk(chunk []byte) (out Chunk) {
	if len(chunk) == 0 {
		return out
	}

	if chunk[0] == ChunkStartMarker {
		out.Start = true
		chunk = chunk[1:]
	}
	if chunk[len(chunk)-1] == ChunkEndMarker {
		out.End = true
		chunk = chunk[:len(chunk)-1]
	}

	out.Data = chunk

	return out
}

// Cut returns a new smaller chunk. NOTE: the options Offset and limit operate
// on *marshalled* data.
func (c Chunk) Cut(opts GetChunkOptions) (Chunk, error) {
	data := c.Marshal()
	size := len(data)

	if opts.Offset > size {
		return Chunk{}, fmt.Errorf("chunk offset greater than size of data: %d > %d", opts.Offset, size)
	}

	// limit cannot be higher than the max
	if opts.Limit > ChunkMaxLimit {
		opts.Limit = ChunkMaxLimit
	}

	// zero means limitless but we set it the size of the remaining data so that
	// it is easier to work with.
	if opts.Limit == 0 {
		opts.Limit = size - opts.Offset
	}

	// Adjust limit if it extends beyond size of value
	if (opts.Offset + opts.Limit) > size {
		opts.Limit = size - opts.Offset
	}

	// Cut data
	data = data[opts.Offset:(opts.Offset + opts.Limit)]

	return UnmarshalChunk(data), nil
}

// Append appends a chunk to an existing chunk
func (c Chunk) Append(chunk Chunk) Chunk {
	c.Data = append(c.Data, chunk.Data...)
	c.Start = chunk.Start
	c.End = chunk.End
	return c
}
