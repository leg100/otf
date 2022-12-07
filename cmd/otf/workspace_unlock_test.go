package main

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/leg100/otf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkspaceUnlock(t *testing.T) {
	org := otf.NewTestOrganization(t)
	ws := otf.NewTestWorkspace(t, org)
	factory := &fakeClientFactory{ws: ws}

	cmd := WorkspaceUnlockCommand(factory)
	cmd.SetArgs([]string{"dev", "--organization", org.Name()})
	got := bytes.Buffer{}
	cmd.SetOut(&got)
	require.NoError(t, cmd.Execute())
	want := fmt.Sprintf("Successfully unlocked workspace %s\n", ws.Name())
	assert.Equal(t, want, got.String())
}

func TestWorkspaceUnlockMissingName(t *testing.T) {
	cmd := WorkspaceUnlockCommand(&fakeClientFactory{})
	cmd.SetArgs([]string{"--organization", "automatize"})
	err := cmd.Execute()
	assert.EqualError(t, err, "accepts 1 arg(s), received 0")
}

func TestWorkspaceUnlockMissingOrganization(t *testing.T) {
	cmd := WorkspaceUnlockCommand(&fakeClientFactory{})
	cmd.SetArgs([]string{"automatize"})
	err := cmd.Execute()
	assert.EqualError(t, err, "required flag(s) \"organization\" not set")
}
