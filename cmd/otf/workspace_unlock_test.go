package main

import (
	"testing"

	"github.com/leg100/otf/http"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkspaceUnlock(t *testing.T) {
	cmd := WorkspaceUnlockCommand(&http.FakeClientFactory{})
	cmd.SetArgs([]string{"dev", "--organization", "automatize"})
	require.NoError(t, cmd.Execute())
}

func TestWorkspaceUnlockMissingName(t *testing.T) {
	cmd := WorkspaceUnlockCommand(&http.FakeClientFactory{})
	cmd.SetArgs([]string{"--organization", "automatize"})
	err := cmd.Execute()
	assert.EqualError(t, err, "accepts 1 arg(s), received 0")
}

func TestWorkspaceUnlockMissingOrganization(t *testing.T) {
	cmd := WorkspaceUnlockCommand(&http.FakeClientFactory{})
	cmd.SetArgs([]string{"automatize"})
	err := cmd.Execute()
	assert.EqualError(t, err, "required flag(s) \"organization\" not set")
}
