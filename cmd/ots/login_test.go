package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoginCommand(t *testing.T) {
	tmpdir := t.TempDir()

	cmd := LoginCommand(FakeHomeDir(tmpdir))
	cmd.SetArgs([]string{"--hostname", "ots.dev:9898"})
	require.NoError(t, cmd.Execute())

	got, err := os.ReadFile(filepath.Join(tmpdir, CredentialsPath))
	require.NoError(t, err)

	want := `{
  "credentials": {
    "ots.dev:9898": {
      "token": "dummy"
    }
  }
}`

	assert.Equal(t, want, string(got))
}

// Test login command doesn't overwrite any existing credentials for TFE etc
func TestLoginCommandWithExistingCredentials(t *testing.T) {
	// Write a config file with existing creds
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

	cmd := LoginCommand(FakeHomeDir(tmpdir))
	cmd.SetArgs([]string{"--hostname", "ots.dev:9898"})
	require.NoError(t, cmd.Execute())

	got, err := os.ReadFile(path)
	require.NoError(t, err)

	want := `{
  "credentials": {
    "app.terraform.io": {
      "token": "secret"
    },
    "ots.dev:9898": {
      "token": "dummy"
    }
  }
}`

	assert.Equal(t, want, string(got))
}

func TestLoginCommandNoHostname(t *testing.T) {
	// Ensure env var doesn't interfere with test
	os.Unsetenv("OTS_HOSTNAME")

	cmd := LoginCommand(nil)
	require.Equal(t, ErrMissingHostname, cmd.Execute())
}

type FakeHomeDir string

func (f FakeHomeDir) UserHomeDir() (string, error) { return string(f), nil }
