/*
Package filestore provides filesystem storage for binary objects (blobs).
*/
package filestore

import (
	"os"
	"path/filepath"

	"github.com/leg100/ots"
)

const (
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
func (fs *FileStore) Get(bid ots.BlobID) ([]byte, error) {
	return os.ReadFile(fs.fpath(bid, false))
}

// GetChunk retrieves a chunk of bytes of the blob.
func (fs *FileStore) GetChunk(bid ots.BlobID, opts ots.GetChunkOptions) ([]byte, error) {
	complete := true

	// Check whether complete or incomplete file exists
	f, err := os.ReadFile(fs.fpath(bid, false))
	if err != nil {
		if os.IsNotExist(err) {
			f, err = os.ReadFile(fs.fpath(bid, true))
			if err != nil {
				return nil, err
			}
			complete = false
		} else {
			return nil, err
		}
	}

	return ots.GetChunk(f, opts, complete)
}

// Put writes a complete blob in one go.
func (fs *FileStore) Put(bid ots.BlobID, p []byte) error {
	return os.WriteFile(fs.fpath(bid, false), p, Perms)
}

// PutChunk writes a chunk of bytes of a blob.
func (fs *FileStore) PutChunk(bid ots.BlobID, chunk []byte, opts ots.PutChunkOptions) error {
	f, err := os.OpenFile(fs.fpath(bid, true), os.O_APPEND, Perms)
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
		if err := os.Link(fs.fpath(bid, true), fs.fpath(bid, false)); err != nil {
			return err
		}
	}

	return nil
}

// Create creates a new blob with the given content. Set chunked=true if further
// chunks are to be written before the blob is deemed complete.
func (fs *FileStore) Create(p []byte, opts ots.CreateBlobOptions) (ots.BlobID, error) {
	bid := ots.NewBlobID()

	if err := os.WriteFile(fs.fpath(bid, opts.Chunked), p, Perms); err != nil {
		return "", err
	}

	return bid, nil
}

func (fs *FileStore) fpath(blob ots.BlobID, incomplete bool) string {
	name := filepath.Join(fs.Path, string(blob))
	if incomplete {
		name = name + ".incomplete"
	}
	return name
}
