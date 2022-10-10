package main

import (
	"bytes"
	"testing"

	"github.com/leg100/otf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkspaceEdit(t *testing.T) {
	org := otf.NewTestOrganization(t)
	ws := otf.NewTestWorkspace(t, org)
	factory := &fakeClientFactory{ws: ws}

	cmd := WorkspaceEditCommand(factory)

	t.Run("help", func(t *testing.T) {
		cmd.SetArgs([]string{"dev", "--organization", org.Name()})
		require.NoError(t, cmd.Execute())
	})

	t.Run("update execution mode", func(t *testing.T) {
		cmd.SetArgs([]string{"dev", "--organization", org.Name(), "--execution-mode", "local"})
		buf := bytes.Buffer{}
		cmd.SetOut(&buf)
		require.NoError(t, cmd.Execute())

		assert.Equal(t, "updated execution mode: local\n", buf.String())
	})
}

func TestWorkspaceEditMissingOrganization(t *testing.T) {
	cmd := WorkspaceEditCommand(&fakeClientFactory{})
	cmd.SetArgs([]string{"automatize"})
	err := cmd.Execute()
	assert.EqualError(t, err, "required flag(s) \"organization\" not set")
}
