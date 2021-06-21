package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkspaceCommand(t *testing.T) {
	cmd := WorkspaceLockCommand(&FakeClientConfig{})
	cmd.SetArgs([]string{"dev", "--organization", "automatize"})
	require.NoError(t, cmd.Execute())
}

func TestWorkspaceCommandMissingName(t *testing.T) {
	cmd := WorkspaceLockCommand(&FakeClientConfig{})
	cmd.SetArgs([]string{"--organization", "automatize"})
	err := cmd.Execute()
	assert.EqualError(t, err, "accepts 1 arg(s), received 0")
}

func TestWorkspaceCommandMissingOrganization(t *testing.T) {
	cmd := WorkspaceLockCommand(&FakeClientConfig{})
	cmd.SetArgs([]string{"automatize"})
	err := cmd.Execute()
	assert.EqualError(t, err, "required flag(s) \"organization\" not set")
}
