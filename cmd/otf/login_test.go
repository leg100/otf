package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoginCommand(t *testing.T) {
	var store KVStore = KVMap(make(map[string]string))

	cmd := LoginCommand(store, "localhost:8080")
	require.NoError(t, cmd.Execute())

	token, _ := store.Load("localhost:8080")
	assert.Equal(t, "dummy", token)
}
