package main

import (
	"testing"

	"github.com/leg100/otf/http"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkspaceList(t *testing.T) {
	cmd := WorkspaceListCommand(&http.FakeClientFactory{})
	cmd.SetArgs([]string{"--organization", "automatize"})
	require.NoError(t, cmd.Execute())
}

func TestWorkspaceListMissingOrganization(t *testing.T) {
	cmd := WorkspaceListCommand(&http.FakeClientFactory{})
	cmd.SetArgs([]string{"automatize"})
	err := cmd.Execute()
	assert.EqualError(t, err, "required flag(s) \"organization\" not set")
}
