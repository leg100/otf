package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCrypto(t *testing.T) {
	secret := GenerateRandomString(32)
	encrypted, err := Encrypt([]byte("exampleplaintext"), secret)
	require.NoError(t, err)

	decrypted, err := Decrypt(encrypted, secret)
	require.NoError(t, err, encrypted)
	assert.Equal(t, "exampleplaintext", string(decrypted))
}
