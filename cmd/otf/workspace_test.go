package main

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/leg100/otf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkspaceEdit(t *testing.T) {
	org := otf.NewTestOrganization(t)
	ws := otf.NewTestWorkspace(t, org)
	app := fakeApp(withFakeWorkspaces(ws))

	t.Run("help", func(t *testing.T) {
		cmd := app.workspaceEditCommand()
		cmd.SetArgs([]string{"dev", "--organization", org.Name()})
		require.NoError(t, cmd.Execute())
	})

	t.Run("update execution mode", func(t *testing.T) {
		cmd := app.workspaceEditCommand()
		cmd.SetArgs([]string{"dev", "--organization", org.Name(), "--execution-mode", "local"})
		buf := bytes.Buffer{}
		cmd.SetOut(&buf)
		require.NoError(t, cmd.Execute())

		assert.Equal(t, "updated execution mode: local\n", buf.String())
	})

	t.Run("missing organization", func(t *testing.T) {
		cmd := app.workspaceEditCommand()
		cmd.SetArgs([]string{"automatize"})
		err := cmd.Execute()
		assert.EqualError(t, err, "required flag(s) \"organization\" not set")
	})
}

func TestWorkspaceShow(t *testing.T) {
	org := otf.NewTestOrganization(t)
	ws := otf.NewTestWorkspace(t, org)
	app := fakeApp(withFakeWorkspaces(ws))

	cmd := app.workspaceShowCommand()
	cmd.SetArgs([]string{"dev", "--organization", "automatize"})
	require.NoError(t, cmd.Execute())

	t.Run("missing organization", func(t *testing.T) {
		cmd := app.workspaceShowCommand()
		cmd.SetArgs([]string{"automatize"})
		err := cmd.Execute()
		assert.EqualError(t, err, "required flag(s) \"organization\" not set")
	})
}

func TestWorkspaceList(t *testing.T) {
	org := otf.NewTestOrganization(t)
	ws1 := otf.NewTestWorkspace(t, org)
	ws2 := otf.NewTestWorkspace(t, org)
	app := fakeApp(withFakeWorkspaces(ws1, ws2))

	cmd := app.workspaceListCommand()
	cmd.SetArgs([]string{"--organization", org.Name()})
	got := bytes.Buffer{}
	cmd.SetOut(&got)
	require.NoError(t, cmd.Execute())
	want := fmt.Sprintf("%s\n%s\n", ws1.Name(), ws2.Name())
	assert.Equal(t, want, got.String())

	t.Run("missing organization", func(t *testing.T) {
		cmd := app.workspaceListCommand()
		cmd.SetArgs([]string{"automatize"})
		err := cmd.Execute()
		assert.EqualError(t, err, "required flag(s) \"organization\" not set")
	})
}

func TestWorkspaceLock(t *testing.T) {
	org := otf.NewTestOrganization(t)
	ws := otf.NewTestWorkspace(t, org)
	app := fakeApp(withFakeWorkspaces(ws))

	cmd := app.workspaceLockCommand()
	cmd.SetArgs([]string{"dev", "--organization", "automatize"})
	got := bytes.Buffer{}
	cmd.SetOut(&got)
	require.NoError(t, cmd.Execute())
	want := fmt.Sprintf("Successfully locked workspace %s\n", ws.Name())
	assert.Equal(t, want, got.String())

	t.Run("missing name", func(t *testing.T) {
		cmd := fakeApp().workspaceLockCommand()
		cmd.SetArgs([]string{"--organization", "automatize"})
		err := cmd.Execute()
		assert.EqualError(t, err, "accepts 1 arg(s), received 0")
	})

	t.Run("missing organization", func(t *testing.T) {
		cmd := fakeApp().workspaceLockCommand()
		cmd.SetArgs([]string{"automatize"})
		err := cmd.Execute()
		assert.EqualError(t, err, "required flag(s) \"organization\" not set")
	})
}

func TestWorkspaceUnlock(t *testing.T) {
	org := otf.NewTestOrganization(t)
	ws := otf.NewTestWorkspace(t, org)
	app := fakeApp(withFakeWorkspaces(ws))

	cmd := app.workspaceUnlockCommand()
	cmd.SetArgs([]string{"dev", "--organization", org.Name()})
	got := bytes.Buffer{}
	cmd.SetOut(&got)
	require.NoError(t, cmd.Execute())
	want := fmt.Sprintf("Successfully unlocked workspace %s\n", ws.Name())
	assert.Equal(t, want, got.String())

	t.Run("missing name", func(t *testing.T) {
		cmd := app.workspaceUnlockCommand()
		cmd.SetArgs([]string{"--organization", "automatize"})
		err := cmd.Execute()
		assert.EqualError(t, err, "accepts 1 arg(s), received 0")
	})

	t.Run("missing organization", func(t *testing.T) {
		cmd := app.workspaceUnlockCommand()
		cmd.SetArgs([]string{"automatize"})
		err := cmd.Execute()
		assert.EqualError(t, err, "required flag(s) \"organization\" not set")
	})
}
