package main

import (
	"testing"

	"github.com/leg100/otf/http"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkspaceEdit(t *testing.T) {
	cmd := WorkspaceEditCommand(&http.FakeClientFactory{})
	cmd.SetArgs([]string{"dev", "--organization", "automatize"})
	require.NoError(t, cmd.Execute())
}

func TestWorkspaceEditMissingOrganization(t *testing.T) {
	cmd := WorkspaceEditCommand(&http.FakeClientFactory{})
	cmd.SetArgs([]string{"automatize"})
	err := cmd.Execute()
	assert.EqualError(t, err, "required flag(s) \"organization\" not set")
}
