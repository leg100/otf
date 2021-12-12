package assets

import (
	"crypto/sha256"
	"fmt"
	"io"
	"io/fs"
	"strings"
)

// StaticFS provides a filesystem for accessing static assets. Handles creating
// and looking up paths containing "cache busting" hashes.
type StaticFS struct {
	// Wrapped is the underlying filesystem.
	wrapped fs.FS
}

// NewStaticFS constructs a StaticFS, wrapping the passed filesystem.
func NewStaticFS(wrap fs.FS) *StaticFS {
	return &StaticFS{wrap}
}

// Open opens the named file after stripping the hash from the name.
func (fs *StaticFS) Open(name string) (fs.File, error) {
	parts := strings.Split(name, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("expected two dots in path: %s", name)
	}

	// new name without hash
	name = fmt.Sprintf("%s.%s", parts[0], parts[1])

	return fs.wrapped.Open(name)
}

// AppendHash inserts a hash of the named file into its filename, before the
// filename extension: <path>.<ext> -> <path>.<hash>.<ext>, where <hash> is the
// hex format of the SHA256 hash of the contents of the file.
func (fs *StaticFS) AppendHash(name string) (string, error) {
	f, err := fs.wrapped.Open(name)
	if err != nil {
		return "", err
	}

	// TODO: this is an expensive operation to perform if this method is to be
	// called everytime a template is rendered; consider caching result.
	h := sha256.New()
	_, err = io.Copy(h, f)
	if err != nil {
		return "", err
	}

	parts := strings.Split(name, ".")
	if len(parts) != 2 {
		return "", fmt.Errorf("expected one dot in path: %s", name)
	}

	nameWithHash := fmt.Sprintf("%s.%s.%s", parts[0], h.Sum(nil), parts[1])

	return nameWithHash, nil
}
