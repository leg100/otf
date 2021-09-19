package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkspaceList(t *testing.T) {
	cmd := WorkspaceListCommand(&FakeClientConfig{})
	cmd.SetArgs([]string{"--organization", "automatize"})
	require.NoError(t, cmd.Execute())
}

func TestWorkspaceListMissingOrganization(t *testing.T) {
	cmd := WorkspaceListCommand(&FakeClientConfig{})
	cmd.SetArgs([]string{"automatize"})
	err := cmd.Execute()
	assert.EqualError(t, err, "required flag(s) \"organization\" not set")
}
