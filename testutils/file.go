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
