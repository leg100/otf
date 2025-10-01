package runner

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

func newWorkdir(workingDirectory string, RunID string) (*workdir, error) {
	// create dedicated directory for run
	// The absolute path needs to be consistent between operations, in order to
	// avoid errors during apply for resources/variables that reference it since
	// the saved plan locks in those values.
	// ref: https://github.com/leg100/otf/pull/822
	rootDir := path.Join(os.TempDir(), "otf-runs", RunID)
	err := os.MkdirAll(rootDir, 0o700)
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
