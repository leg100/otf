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

// String returns the absolute path to the working directory.
func (w *workdir) String() string {
	return path.Join(w.root, w.relative)
}

// writeFile writes a file to the working directory.
func (w *workdir) writeFile(path string, b []byte) error {
	return os.WriteFile(filepath.Join(w.String(), path), b, 0o644)
}

// readFile reads a file from the working directory.
func (w *workdir) readFile(path string) ([]byte, error) {
	return os.ReadFile(filepath.Join(w.String(), path))
}

// close removes working directory
func (w *workdir) close() error {
	return os.RemoveAll(w.root)
}
