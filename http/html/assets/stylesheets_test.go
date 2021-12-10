package assets

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCacheBustingPaths(t *testing.T) {
	fs := newTestFilesystem(t,
		"/test/a.txt", "abc",
		"/test/b.txt", "def",
		"/test/c.txt", "ghi",
	)
	want := []string{
		"test/a.txt?v=ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad",
		"test/b.txt?v=cb8379ac2098aa165029e3938a51da0bcecfc008fd6795f401178647f96c5b34",
		"test/c.txt?v=50ae61e841fac4e8f9e40baf2ad36ec868922ea48368c18f9535e47db56dd7fb",
	}

	paths, err := CacheBustingPaths(fs, "**/*.txt")
	require.NoError(t, err)
	assert.Equal(t, want, paths)
}

// newTestFilesystem creates a temporary filesystem consisting of paths of files
// each with the given contents.
func newTestFilesystem(t *testing.T, pathsAndContents ...string) fs.FS {
	if len(pathsAndContents)%2 != 0 {
		t.Fatal("must provide even number of args")
	}

	dir := t.TempDir()

	for i := 0; i < len(pathsAndContents); i += 2 {
		path := filepath.Join(dir, pathsAndContents[i])

		require.NoError(t, os.MkdirAll(filepath.Dir(path), 0755))
		require.NoError(t, os.WriteFile(path, []byte(pathsAndContents[i+1]), 0644))
	}

	return os.DirFS(dir)
}
