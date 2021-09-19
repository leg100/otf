/*
Package filestore provides filesystem storage for binary objects (blobs).
*/
package filestore

import (
	"os"
	"path/filepath"

	"github.com/leg100/otf"
)

const (
	// Chmod perms for a file blob
	Perms = 0644
)

var _ otf.BlobStore = (*FileStore)(nil)

// FileStore is a filesystem based blob database
type FileStore struct {
	Path string
}

// NewFilestore constructs a filestore rooted at the given path.
func NewFilestore(path string) (*FileStore, error) {
	// Empty path defaults to a temp dir
	if path == "" {
		var err error
		path, err = os.MkdirTemp("", "otf-filestore-")
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
func (fs *FileStore) Get(bid string) ([]byte, error) {
	return os.ReadFile(fs.fpath(bid, false))
}

// GetChunk retrieves a chunk of bytes of the blob.
func (fs *FileStore) GetChunk(bid string, opts otf.GetChunkOptions) ([]byte, error) {
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

	return otf.GetChunk(f, opts, complete)
}

// Put writes a complete blob in one go.
func (fs *FileStore) Put(bid string, p []byte) error {
	return os.WriteFile(fs.fpath(bid, false), p, Perms)
}

// PutChunk writes a chunk of bytes of a blob.
func (fs *FileStore) PutChunk(bid string, chunk []byte, opts otf.PutChunkOptions) error {
	f, err := os.OpenFile(fs.fpath(bid, true), os.O_CREATE|os.O_APPEND|os.O_WRONLY, Perms)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.Write(chunk); err != nil {
		return err
	}

	// Is last chunk?
	if opts.End {
		// Must close file before moving it
		if err := f.Close(); err != nil {
			return err
		}

		// blob.incomplete -> blob
		if err := os.Rename(fs.fpath(bid, true), fs.fpath(bid, false)); err != nil {
			return err
		}
	}

	return nil
}

func (fs *FileStore) fpath(bid string, incomplete bool) string {
	name := filepath.Join(fs.Path, bid)
	if incomplete {
		name = name + ".incomplete"
	}
	return name
}
