package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCredentialsStore(t *testing.T) {
	tmpdir := t.TempDir()

	store, err := NewCredentialsStore(FakeDirectories(tmpdir))
	require.NoError(t, err)

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
	tmpdir := t.TempDir()

	path := filepath.Join(tmpdir, CredentialsPath)
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0755))
	require.NoError(t, os.WriteFile(path, []byte(existing), 0600))

	store, err := NewCredentialsStore(FakeDirectories(tmpdir))
	require.NoError(t, err)

	require.NoError(t, store.Save("otf.dev:8080", "dummy"))

	got, err := os.ReadFile(path)
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

func TestCredentialsStoreSanitizeAddress(t *testing.T) {
	tests := []struct {
		name    string
		address string
		want    string
	}{
		{
			name:    "no scheme",
			address: "localhost:8080",
			want:    "localhost:8080",
		},
		{
			name:    "has scheme",
			address: "https://localhost:8080",
			want:    "localhost:8080",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpdir := t.TempDir()

			store, err := NewCredentialsStore(FakeDirectories(tmpdir))
			require.NoError(t, err)

			address, err := store.sanitizeHostname(tt.address)
			require.NoError(t, err)
			assert.Equal(t, tt.want, address)
		})
	}
}

type FakeDirectories string

func (f FakeDirectories) UserHomeDir() (string, error) { return string(f), nil }
