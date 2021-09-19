package otf

import (
	"fmt"

	"github.com/google/uuid"
)

const (
	// ChunkMaxLimit is maximum permissible size of a chunk
	ChunkMaxLimit = 65536

	// ChunkStartMarker is the special byte that prefixes the first chunk
	ChunkStartMarker = byte(2)

	// ChunkEndMarker is the special byte that suffixes the last chunk
	ChunkEndMarker = byte(3)
)

// BlobStore implementations provide a persistent store from and to which binary
// objects can be fetched and uploaded.
type BlobStore interface {
	// Get fetches a blob
	Get(string) ([]byte, error)

	// Get fetches a blob chunk
	GetChunk(string, GetChunkOptions) ([]byte, error)

	// Put uploads a blob
	Put(string, []byte) error

	// Put uploads a blob chunk
	PutChunk(string, []byte, PutChunkOptions) error
}

type GetChunkOptions struct {
	// The maximum number of bytes of logs to return to the client
	Limit int `schema:"limit"`

	// The start position in the logs from which to send to the client
	Offset int `schema:"offset"`
}

type PutChunkOptions struct {
	// End indicates this is the last and final chunk
	End bool `schema:"end"`
}

// NewBlobID generates a unique blob ID
func NewBlobID() string {
	return uuid.NewString()
}

// GetChunk retrieves a chunk of bytes from a byte slice. The first chunk in the
// slice is prefixed with a special byte. If complete is true then the last
// chunk in the slice is suffixed with a special byte.
func GetChunk(p []byte, opts GetChunkOptions, complete bool) ([]byte, error) {
	p = append([]byte{ChunkStartMarker}, p...)

	if complete {
		p = append(p, ChunkEndMarker)
	}

	if opts.Offset > len(p) {
		return nil, fmt.Errorf("offset greater than size of binary object: %d > %d", opts.Offset, len(p))
	}

	if opts.Limit > ChunkMaxLimit {
		opts.Limit = ChunkMaxLimit
	}

	// Adjust limit if it extends beyond size of binary object
	if (opts.Offset + opts.Limit) > len(p) {
		opts.Limit = len(p) - opts.Offset
	}

	return p[opts.Offset:(opts.Offset + opts.Limit)], nil
}
