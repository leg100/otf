package agent

import (
	"os"
	"path"
	"path/filepath"
)

// workdir is a workspace's working directory
type workdir struct {
	root     string // absolute path of the root directory containing tf config
	relative string // relative path to working directory
}

func newWorkdir(workingDirectory string) (*workdir, error) {
	// create dedicated directory for environment
	rootDir, err := os.MkdirTemp("", "otf-config-")
	if err != nil {
		return nil, err
	}
	return &workdir{
		root:     rootDir,
		relative: workingDirectory,
	}, nil
}

// WriteFile writes a file to the working directory.
func (w *workdir) WriteFile(path string, b []byte) error {
	return os.WriteFile(filepath.Join(w.String(), path), b, 0o644)
}

// ReadFile reads a file from the working directory.
func (w *workdir) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(filepath.Join(w.String(), path))
}

// Close removes working directory
func (w *workdir) Close() error {
	return os.RemoveAll(w.root)
}

// String returns the absolute path to the working directory.
func (w *workdir) String() string {
	return path.Join(w.root, w.relative)
}
