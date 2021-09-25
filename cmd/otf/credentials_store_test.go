package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCredentialsStore(t *testing.T) {
	store := CredentialsStore(filepath.Join(t.TempDir(), "creds.json"))

	require.NoError(t, store.Save("otf.dev:8080", "dummy"))

	token, err := store.Load("otf.dev:8080")
	require.NoError(t, err)

	assert.Equal(t, "dummy", token)
}

func TestCredentialsStoreWithExistingCredentials(t *testing.T) {
	// Write a config file with existing creds for TFC
	existing := `{
     "credentials": {
       "app.terraform.io": {
         "token": "secret"
       }
     }
   }
`

	store := CredentialsStore(filepath.Join(t.TempDir(), "creds.json"))
	require.NoError(t, os.WriteFile(string(store), []byte(existing), 0600))

	require.NoError(t, store.Save("otf.dev:8080", "dummy"))

	got, err := os.ReadFile(string(store))
	require.NoError(t, err)

	want := `{
  "credentials": {
    "app.terraform.io": {
      "token": "secret"
    },
    "otf.dev:8080": {
      "token": "dummy"
    }
  }
}`

	assert.Equal(t, want, string(got))
}
