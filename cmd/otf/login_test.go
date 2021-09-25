package main

import (
	"os"
	"testing"

	"github.com/leg100/otf/http"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoginCommand(t *testing.T) {
	var store http.KVStore = KVMap(make(map[string]string))

	cmd := LoginCommand(store)
	require.NoError(t, cmd.Execute())

	token, _ := store.Load("localhost:8080")
	assert.Equal(t, "dummy", token)
}

func TestLoginCommandWithExplicitAddress(t *testing.T) {
	// Ensure env var doesn't interfere with test
	os.Unsetenv("OTF_ADDRESS")

	var store http.KVStore = KVMap(make(map[string]string))

	cmd := LoginCommand(store)
	cmd.SetArgs([]string{"--address", "otf.dev:8080"})
	require.NoError(t, cmd.Execute())

	token, _ := store.Load("otf.dev:8080")
	assert.Equal(t, "dummy", token)
}
