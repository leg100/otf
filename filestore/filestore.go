/*
Package filestore provides filesystem storage for binary objects (blobs).
*/
package filestore

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/leg100/ots"
)

const (
	MaxLimit    = 65536
	StartMarker = byte(2)
	EndMarker   = byte(3)

	// Chmod perms for a file blob
	Perms = 0644
)

var _ ots.BlobStore = (*FileStore)(nil)

// FileStore is a filesystem based blob database
type FileStore struct {
	Path string
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

	return &FileStore{Path: path}, nil
}

// Get retrieves a complete blob.
func (fs *FileStore) Get(b ots.Blob) ([]byte, error) {
	return os.ReadFile(fs.fpath(b, false))
}

// GetChunk retrieves a chunk of bytes of the blob.
func (fs *FileStore) GetChunk(b ots.Blob, opts ots.GetBlobOptions) ([]byte, error) {
	completed := true

	// Check whether complete or incomplete file exists
	f, err := os.ReadFile(fs.fpath(b, false))
	if err != nil {
		if os.IsNotExist(err) {
			f, err = os.ReadFile(fs.fpath(b, true))
			if err != nil {
				return nil, err
			}
			completed = false
		} else {
			return nil, err
		}
	}

	if opts.Offset == 0 {
		f = append([]byte{StartMarker}, f...)
	}

	if completed {
		f = append(f, EndMarker)
	}

	if opts.Offset > len(f) {
		return nil, fmt.Errorf("offset greater than size of binary object")
	}

	if opts.Limit > MaxLimit {
		opts.Limit = MaxLimit
	}

	// Adjust limit if it extends beyond size of binary object
	if (opts.Offset + opts.Limit) > len(f) {
		opts.Limit = len(f) - opts.Offset
	}

	return f[opts.Offset:(opts.Offset + opts.Limit)], nil
}

// Put writes a complete blob in one go.
func (fs *FileStore) Put(blob ots.Blob, p []byte) error {
	return os.WriteFile(fs.fpath(blob, false), p, Perms)
}

// PutChunk writes a chunk of bytes of a blob.
func (fs *FileStore) PutChunk(b ots.Blob, chunk []byte, opts ots.PutBlobOptions) error {
	f, err := os.OpenFile(fs.fpath(b, true), os.O_APPEND, Perms)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.Write(chunk); err != nil {
		return err
	}

	// Is last chunk?
	if opts.End {
		f.Close() // Must close file before moving it

		// blob.incomplete -> blob
		if err := os.Link(fs.fpath(b, true), fs.fpath(b, false)); err != nil {
			return err
		}
	}

	return nil
}

// Create creates a new blob with the given content. Set chunked=true if further
// chunks are to be written before the blob is complete.
func (fs *FileStore) Create(p []byte, opts ots.CreateBlobOptions) (ots.Blob, error) {
	blob := ots.Blob(newID())

	if err := os.WriteFile(fs.fpath(blob, opts.Chunked), p, Perms); err != nil {
		return "", err
	}

	return blob, nil
}

func (fs *FileStore) fpath(blob ots.Blob, incomplete bool) string {
	name := filepath.Join(fs.Path, string(blob))
	if incomplete {
		name = name + ".incomplete"
	}
	return name
}

// Generate a new unique ID for a filestore blob
func newID() string {
	return uuid.NewString()
}
