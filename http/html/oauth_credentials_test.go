package html

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOAuthCredentials_Valid(t *testing.T) {
	creds := &OAuthCredentials{
		prefix:       "fake",
		clientID:     "abc123",
		clientSecret: "xyz789",
	}
	assert.NoError(t, creds.Valid())
}

func TestOAuthCredentials_Incomplete(t *testing.T) {
	onlyClientID := &OAuthCredentials{
		prefix:   "fake",
		clientID: "abc123",
	}
	assert.Equal(t, ErrOAuthCredentialsIncomplete, onlyClientID.Valid())

	onlyClientSecret := &OAuthCredentials{
		prefix:       "fake",
		clientSecret: "xyz789",
	}
	assert.Equal(t, ErrOAuthCredentialsIncomplete, onlyClientSecret.Valid())
}

func TestOAuthCredentials_Unspecified(t *testing.T) {
	unspecified := &OAuthCredentials{
		prefix: "fake",
	}
	assert.Equal(t, ErrOAuthCredentialsUnspecified, unspecified.Valid())
}
