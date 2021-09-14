package ots

// BlobStore implementations provide a persistent store from and to which binary
// objects can be fetched and uploaded.
type BlobStore interface {
	// Get fetches a blob
	Get(Blob) ([]byte, error)

	// Get fetches a blob chunk
	GetChunk(Blob, GetBlobOptions) ([]byte, error)

	// Put uploads a blob
	Put(Blob, []byte) error

	// Put uploads a blob chunk
	PutChunk(Blob, []byte, PutBlobOptions) error

	// Create creates a new blob
	Create([]byte, CreateBlobOptions) (Blob, error)
}

type Blob string

type GetBlobOptions struct {
	// The maximum number of bytes of logs to return to the client
	Limit int `schema:"limit"`

	// The start position in the logs from which to send to the client
	Offset int `schema:"offset"`
}

type PutBlobOptions struct {
	// Start indicates this is the first chunk
	Start bool `schema:"start"`

	// End indicates this is the last and final chunk
	End bool `schema:"end"`
}

type CreateBlobOptions struct {
	// Chunked is whether the blob is split into chunks.
	Chunked bool `schema:"chunked"`
}
