package testutils

import (
	"testing"

	"crypto/rand"

	"github.com/stretchr/testify/require"
)

func NewSecret(t *testing.T) []byte {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	require.NoError(t, err)
	return b
}
