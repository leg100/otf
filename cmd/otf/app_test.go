package main

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetToken_Env(t *testing.T) {
	t.Setenv("OTF_TOKEN", "mytoken")
	got, err := (&application{}).getToken("localhost:8080")
	require.NoError(t, err)
	assert.Equal(t, "mytoken", got)
}

func TestSetToken_HostSpecificEnv(t *testing.T) {
	t.Setenv("TF_TOKEN_otf_dev", "mytoken")
	got, err := (&application{}).getToken("otf.dev")
	require.NoError(t, err)
	assert.Equal(t, "mytoken", got)
}

func TestSetToken_CredentialStore(t *testing.T) {
	store := CredentialsStore(filepath.Join(t.TempDir(), "creds.json"))
	require.NoError(t, store.Save("otf.dev", "mytoken"))

	got, err := (&application{creds: store}).getToken("otf.dev")
	require.NoError(t, err)
	assert.Equal(t, "mytoken", got)
}
