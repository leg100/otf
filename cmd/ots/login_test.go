package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoginCommand(t *testing.T) {
	tmpdir := t.TempDir()

	cmd := LoginCommand(FakeDirectories(tmpdir))
	require.NoError(t, cmd.Execute())

	store, err := NewCredentialsStore(FakeDirectories(tmpdir))
	require.NoError(t, err)
	token, err := store.Load("localhost:8080")
	require.NoError(t, err)
	assert.Equal(t, "dummy", token)
}

func TestLoginCommandWithExplicitAddress(t *testing.T) {
	// Ensure env var doesn't interfere with test
	os.Unsetenv("OTS_ADDRESS")

	tmpdir := t.TempDir()

	cmd := LoginCommand(FakeDirectories(tmpdir))
	cmd.SetArgs([]string{"--address", "ots.dev:8080"})
	require.NoError(t, cmd.Execute())

	store, err := NewCredentialsStore(FakeDirectories(tmpdir))
	require.NoError(t, err)
	_, err = store.Load("ots.dev:8080")
	require.NoError(t, err)
}
