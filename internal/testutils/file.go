// Package testutils provides test helpers.
package testutils

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func ReadFile(t *testing.T, path string) []byte {
	t.Helper()

	contents, err := os.ReadFile(path)
	require.NoError(t, err)
	return contents
}

func TempFile(t *testing.T, data []byte) (path string) {
	t.Helper()

	f, err := os.CreateTemp(os.TempDir(), "")
	require.NoError(t, err)

	t.Cleanup(func() {
		os.Remove(f.Name())
	})

	_, err = f.Write(data)
	require.NoError(t, err)

	return f.Name()
}
