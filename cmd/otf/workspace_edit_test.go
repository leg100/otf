package main

import (
	"bytes"
	"testing"

	"github.com/leg100/otf"
	"github.com/leg100/otf/http"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkspaceEdit(t *testing.T) {
	org, err := otf.NewOrganization(otf.OrganizationCreateOptions{Name: otf.String("automatize")})
	require.NoError(t, err)
	ws, err := otf.NewWorkspace(org, otf.WorkspaceCreateOptions{Name: "dev"})
	require.NoError(t, err)
	factory := &http.FakeClientFactory{Workspace: ws}

	cmd := WorkspaceEditCommand(factory)

	t.Run("help", func(t *testing.T) {
		cmd.SetArgs([]string{"dev", "--organization", "automatize"})
		require.NoError(t, cmd.Execute())
	})

	t.Run("update execution mode", func(t *testing.T) {
		cmd.SetArgs([]string{"dev", "--organization", "automatize", "--execution-mode", "local"})
		buf := bytes.Buffer{}
		cmd.SetOut(&buf)
		require.NoError(t, cmd.Execute())

		assert.Equal(t, "updated execution mode: local\n", buf.String())
	})
}

func TestWorkspaceEditMissingOrganization(t *testing.T) {
	cmd := WorkspaceEditCommand(&http.FakeClientFactory{})
	cmd.SetArgs([]string{"automatize"})
	err := cmd.Execute()
	assert.EqualError(t, err, "required flag(s) \"organization\" not set")
}
