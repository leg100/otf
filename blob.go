package otf

import (
	"github.com/google/uuid"
)

// BlobStore implementations provide a persistent store from and to which binary
// objects can be fetched and uploaded.
type BlobStore interface {
	// Get fetches a blob
	Get(string) ([]byte, error)

	// Put uploads a blob
	Put(string, []byte) error
}

func NewBlobID() string {
	return uuid.NewString()
}
