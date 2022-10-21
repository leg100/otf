package main

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/leg100/otf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkspaceLock(t *testing.T) {
	org := otf.NewTestOrganization(t)
	ws := otf.NewTestWorkspace(t, org, otf.WorkspaceCreateOptions{})
	factory := &fakeClientFactory{ws: ws}

	cmd := WorkspaceLockCommand(factory)
	cmd.SetArgs([]string{"dev", "--organization", "automatize"})
	got := bytes.Buffer{}
	cmd.SetOut(&got)
	require.NoError(t, cmd.Execute())
	want := fmt.Sprintf("Successfully locked workspace %s\n", ws.Name())
	assert.Equal(t, want, got.String())
}

func TestWorkspaceLockMissingName(t *testing.T) {
	cmd := WorkspaceLockCommand(&fakeClientFactory{})
	cmd.SetArgs([]string{"--organization", "automatize"})
	err := cmd.Execute()
	assert.EqualError(t, err, "accepts 1 arg(s), received 0")
}

func TestWorkspaceLockMissingOrganization(t *testing.T) {
	cmd := WorkspaceLockCommand(&fakeClientFactory{})
	cmd.SetArgs([]string{"automatize"})
	err := cmd.Execute()
	assert.EqualError(t, err, "required flag(s) \"organization\" not set")
}
