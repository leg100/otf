/*
Package filestore provides filesystem storage for binary objects (blobs).
*/
package filestore

import (
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/leg100/ots"
)

var _ ots.BlobStore = (*FileStore)(nil)

// FileStore is a filesystem based blob database
type FileStore struct {
	path string
}

// NewFilestore constructs a filestore rooted at the given path.
func NewFilestore(path string) (*FileStore, error) {
	// Empty path defaults to a temp dir
	if path == "" {
		var err error
		path, err = os.MkdirTemp("", "ots-filestore-")
		if err != nil {
			return nil, err
		}
	}

	// Ensure path exists
	if err := os.MkdirAll(path, 0755); err != nil {
		return nil, err
	}

	// Ensure path is accessible (MkdirAll won't set perms if path already
	// exists)
	if err := os.Chmod(path, 0755); err != nil {
		return nil, err
	}

	return &FileStore{path: path}, nil
}

func (fs *FileStore) Get(id string) ([]byte, error) {
	return os.ReadFile(filepath.Join(fs.path, id))
}

func (fs *FileStore) Put(blob []byte) (string, error) {
	id := newID()

	if err := os.WriteFile(filepath.Join(fs.path, id), blob, 0644); err != nil {
		return "", err
	}

	return id, nil
}

func (fs *FileStore) Path() string {
	return fs.path
}

// Generate a new unique ID for a filestore blob
func newID() string {
	return uuid.NewString()
}
