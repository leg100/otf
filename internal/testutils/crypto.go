package testutils

import (
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/require"
)

// NewSecret produces a cryptographically random 16-byte key, intended for use
// with aes-128 in tests.
func NewSecret(t *testing.T) []byte {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	require.NoError(t, err)
	return b
}
