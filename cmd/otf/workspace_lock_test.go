package main

import (
	"testing"

	"github.com/leg100/otf"
	"github.com/leg100/otf/http"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkspaceLock(t *testing.T) {
	org, err := otf.NewOrganization(otf.OrganizationCreateOptions{Name: otf.String("automatize")})
	require.NoError(t, err)
	ws, err := otf.NewWorkspace(org, otf.WorkspaceCreateOptions{Name: "dev"})
	require.NoError(t, err)
	factory := &http.FakeClientFactory{Workspace: ws}

	cmd := WorkspaceLockCommand(factory)
	cmd.SetArgs([]string{"dev", "--organization", "automatize"})
	require.NoError(t, cmd.Execute())
}

func TestWorkspaceLockMissingName(t *testing.T) {
	cmd := WorkspaceLockCommand(&http.FakeClientFactory{})
	cmd.SetArgs([]string{"--organization", "automatize"})
	err := cmd.Execute()
	assert.EqualError(t, err, "accepts 1 arg(s), received 0")
}

func TestWorkspaceLockMissingOrganization(t *testing.T) {
	cmd := WorkspaceLockCommand(&http.FakeClientFactory{})
	cmd.SetArgs([]string{"automatize"})
	err := cmd.Execute()
	assert.EqualError(t, err, "required flag(s) \"organization\" not set")
}
