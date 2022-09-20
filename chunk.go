package otf

import (
	"context"
	"fmt"
	"strconv"
)

// Chunk is a section of logs.
type Chunk struct {
	// ID of run that generated the chunk
	RunID string `schema:"run_id,required"`
	// Phase that generated the chunk
	Phase PhaseType `schema:"phase,required"`
	// Position within logs.
	Offset int `schema:"offset,required"`
	// The chunk of logs
	Data []byte
}

// Cut returns a new, smaller chunk.
func (c Chunk) Cut(opts GetChunkOptions) Chunk {
	if opts.Offset > c.NextOffset() {
		// offset is out of bounds - return an empty chunk with offset set to
		// the end of the chunk
		return Chunk{Offset: c.NextOffset()}
	}
	// sanitize limit - 0 means limitless.
	if (opts.Offset+opts.Limit) > c.NextOffset() || opts.Limit == 0 {
		opts.Limit = c.NextOffset() - opts.Offset
	}

	c.Data = c.Data[(opts.Offset - c.Offset):((opts.Offset - c.Offset) + opts.Limit)]
	c.Offset = opts.Offset

	return c
}

func (c Chunk) NextOffset() int {
	return c.Offset + len(c.Data)
}

func (c Chunk) AddStartMarker() Chunk {
	c.Data = append([]byte{0x02}, c.Data...)
	return c
}

func (c Chunk) RemoveStartMarker() Chunk {
	if c.IsStart() {
		c.Data = c.Data[1:]
		c.Offset++
	}
	return c
}

func (c Chunk) AddEndMarker() Chunk {
	c.Data = append(c.Data, 0x03)
	return c
}

func (c Chunk) RemoveEndMarker() Chunk {
	if c.IsEnd() {
		c.Data = c.Data[:len(c.Data)-1]
	}
	return c
}

func (c Chunk) IsStart() bool {
	return len(c.Data) > 0 && c.Data[0] == 0x02
}

func (c Chunk) IsEnd() bool {
	return len(c.Data) > 0 && c.Data[len(c.Data)-1] == 0x03
}

func (c Chunk) Key() string {
	return fmt.Sprintf("%s.%s.log", c.RunID, string(c.Phase))
}

// PersistedChunk is a chunk that has been persisted to the backend.
type PersistedChunk struct {
	// ChunkID uniquely identifies the chunk.
	ChunkID int
	Chunk
}

func (c PersistedChunk) ID() string     { return strconv.Itoa(c.ChunkID) }
func (c PersistedChunk) String() string { return strconv.Itoa(c.ChunkID) }

// LogService is an alias for ChunkService
type LogService ChunkService

// ChunkService provides interaction with chunks.
type ChunkService interface {
	// GetChunk fetches a chunk.
	GetChunk(ctx context.Context, opts GetChunkOptions) (Chunk, error)
	// PutChunk uploads a chunk.
	PutChunk(ctx context.Context, chunk Chunk) error
}

// ChunkStore implementations provide a persistent store from and to which chunks
// can be fetched and uploaded.
type ChunkStore interface {
	// GetChunk fetches a chunk of logs.
	GetChunk(ctx context.Context, opts GetChunkOptions) (Chunk, error)
	// GetChunkByID fetches a specific chunk with the given ID.
	GetChunkByID(ctx context.Context, id int) (PersistedChunk, error)
	// PutChunk uploads a chunk, receiving back the chunk along with a unique
	// ID.
	PutChunk(ctx context.Context, chunk Chunk) (PersistedChunk, error)
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

type PutChunkOptions struct {
	// Start indicates this is the first chunk
	Start bool `schema:"start"`
	// End indicates this is the last and final chunk
	End bool `schema:"end"`
}
