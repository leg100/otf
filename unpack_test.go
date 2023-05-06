package internal

import (
	"bytes"
	"os"
	"path"
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
	filepath.Walk(dst, func(path string, _ os.FileInfo, _ error) error {
		path, err := filepath.Rel(dst, path)
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

// TestUnpack_Github tests unpacking a Github archive of a git repository. Their
// archive is somewhat idiosyncratic in that it uses the PAX format, with
// key-values embedded in the tar file. It's tripped up the unpack code before,
// so putting this test in here to prevent a regression.
func TestUnpack_Github(t *testing.T) {
	tarball, err := os.Open("testdata/github.tar.gz")
	require.NoError(t, err)

	require.NoError(t, Unpack(tarball, t.TempDir()))
}

func TestPack(t *testing.T) {
	tarball, err := Pack("testdata/pack")
	require.NoError(t, err)

	// unpack tarball to test its contents
	dst := t.TempDir()
	err = Unpack(bytes.NewReader(tarball), dst)
	require.NoError(t, err)

	assert.FileExists(t, path.Join(dst, "file"))
	assert.DirExists(t, path.Join(dst, "dir"))
	symlink, err := os.Readlink(path.Join(dst, "dir", "symlink"))
	if assert.NoError(t, err) {
		assert.Equal(t, "../file", symlink)
	}
	assert.FileExists(t, path.Join(dst, "dir", "file"))
}
