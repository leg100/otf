package main

import (
	"testing"

	"github.com/leg100/otf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoginCommand(t *testing.T) {
	var store KVStore = KVMap(make(map[string]string))

	cmd := LoginCommand(store, otf.String("localhost:8080"))
	require.NoError(t, cmd.Execute())

	token, _ := store.Load("localhost:8080")
	assert.Equal(t, "dummy", token)
}

func TestLoginCommand_NonDefaultAddress(t *testing.T) {
	var store KVStore = KVMap(make(map[string]string))

	cmd := LoginCommand(store, otf.String("devd.io:8000"))
	require.NoError(t, cmd.Execute())

	token, _ := store.Load("devd.io:8000")
	assert.Equal(t, "dummy", token)
}
