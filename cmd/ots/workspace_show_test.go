package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkspaceShow(t *testing.T) {
	cmd := WorkspaceShowCommand(&FakeClientConfig{})
	cmd.SetArgs([]string{"dev", "--organization", "automatize"})
	require.NoError(t, cmd.Execute())
}

func TestWorkspaceShowMissingOrganization(t *testing.T) {
	cmd := WorkspaceShowCommand(&FakeClientConfig{})
	cmd.SetArgs([]string{"automatize"})
	err := cmd.Execute()
	assert.EqualError(t, err, "required flag(s) \"organization\" not set")
}
