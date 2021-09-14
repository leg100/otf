package ots

import "github.com/google/uuid"

// BlobID is a binary object identifier
type BlobID string

// BlobStore implementations provide a persistent store from and to which binary
// objects can be fetched and uploaded.
type BlobStore interface {
	// Get fetches a blob
	Get(BlobID) ([]byte, error)

	// Get fetches a blob chunk
	GetChunk(BlobID, GetChunkOptions) ([]byte, error)

	// Put uploads a blob
	Put(BlobID, []byte) error

	// Put uploads a blob chunk
	PutChunk(BlobID, []byte, PutChunkOptions) error

	// Create creates a new blob
	Create([]byte, CreateBlobOptions) (BlobID, error)
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

type CreateBlobOptions struct {
	// Chunked is whether the blob is split into chunks.
	Chunked bool `schema:"chunked"`
}

// NewBlobID generates a unique blob ID
func NewBlobID() BlobID {
	return BlobID(uuid.NewString())
}
