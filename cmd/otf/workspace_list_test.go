package main

import (
	"testing"

	"github.com/leg100/otf"
	"github.com/leg100/otf/http"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkspaceList(t *testing.T) {
	org, err := otf.NewOrganization(otf.OrganizationCreateOptions{Name: otf.String("automatize")})
	require.NoError(t, err)
	ws, err := otf.NewWorkspace(org, otf.WorkspaceCreateOptions{Name: "dev"})
	require.NoError(t, err)
	factory := &http.FakeClientFactory{Workspace: ws}

	cmd := WorkspaceListCommand(factory)
	cmd.SetArgs([]string{"--organization", "automatize"})
	require.NoError(t, cmd.Execute())
}

func TestWorkspaceListMissingOrganization(t *testing.T) {
	cmd := WorkspaceListCommand(&http.FakeClientFactory{})
	cmd.SetArgs([]string{"automatize"})
	err := cmd.Execute()
	assert.EqualError(t, err, "required flag(s) \"organization\" not set")
}
