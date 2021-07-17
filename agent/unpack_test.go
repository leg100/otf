package agent

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnpack(t *testing.T) {
	dst := t.TempDir()

	tarball, err := os.Open("testdata/unpack.tar.gz")
	require.NoError(t, err)

	require.NoError(t, Unpack(tarball, dst))

	var got []string
	filepath.Walk(dst, func(path string, info os.FileInfo, err error) error {
		path, err = filepath.Rel(dst, path)
		require.NoError(t, err)
		got = append(got, path)
		return nil
	})
	assert.Equal(t, []string{
		".",
		"dir",
		"dir/file",
		"dir/symlink",
		"file",
	}, got)
}
