package workspace

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"testing"

	"github.com/leg100/otf/internal/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkspaceEdit(t *testing.T) {
	ws := &Workspace{}
	app := newFakeCLI(ws)

	t.Run("help", func(t *testing.T) {
		cmd := app.workspaceEditCommand()
		cmd.SetArgs([]string{"dev", "--organization", "acme-corp"})
		require.NoError(t, cmd.Execute())
	})

	t.Run("update execution mode", func(t *testing.T) {
		cmd := app.workspaceEditCommand()
		cmd.SetArgs([]string{"dev", "--organization", "acme-corp", "--execution-mode", "local"})
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
	ws := &Workspace{ID: "ws-123"}
	app := newFakeCLI(ws)

	cmd := app.workspaceShowCommand()
	cmd.SetArgs([]string{"dev", "--organization", "automatize"})
	cmd.SetOut(io.Discard)
	require.NoError(t, cmd.Execute())

	t.Run("missing organization", func(t *testing.T) {
		cmd := app.workspaceShowCommand()
		cmd.SetArgs([]string{"automatize"})
		err := cmd.Execute()
		assert.EqualError(t, err, "required flag(s) \"organization\" not set")
	})
}

func TestWorkspaceList(t *testing.T) {
	ws1 := &Workspace{ID: "ws-123"}
	ws2 := &Workspace{ID: "ws-123"}
	app := newFakeCLI(ws1, ws2)

	cmd := app.workspaceListCommand()
	cmd.SetArgs([]string{"--organization", "acme-corp"})
	got := bytes.Buffer{}
	cmd.SetOut(&got)
	require.NoError(t, cmd.Execute())
	want := fmt.Sprintf("%s\n%s\n", ws1.Name, ws2.Name)
	assert.Equal(t, want, got.String())

	t.Run("missing organization", func(t *testing.T) {
		cmd := app.workspaceListCommand()
		cmd.SetArgs([]string{"automatize"})
		err := cmd.Execute()
		assert.EqualError(t, err, "required flag(s) \"organization\" not set")
	})
}

func TestWorkspaceLock(t *testing.T) {
	ws := &Workspace{ID: "ws-123"}
	app := newFakeCLI(ws)

	cmd := app.workspaceLockCommand()
	cmd.SetArgs([]string{"dev", "--organization", "automatize"})
	got := bytes.Buffer{}
	cmd.SetOut(&got)
	require.NoError(t, cmd.Execute())
	want := fmt.Sprintf("Successfully locked workspace %s\n", ws.Name)
	assert.Equal(t, want, got.String())

	t.Run("missing name", func(t *testing.T) {
		cmd := newFakeCLI().workspaceLockCommand()
		cmd.SetArgs([]string{"--organization", "automatize"})
		err := cmd.Execute()
		assert.EqualError(t, err, "accepts 1 arg(s), received 0")
	})

	t.Run("missing organization", func(t *testing.T) {
		cmd := newFakeCLI().workspaceLockCommand()
		cmd.SetArgs([]string{"automatize"})
		err := cmd.Execute()
		assert.EqualError(t, err, "required flag(s) \"organization\" not set")
	})
}

func TestWorkspaceUnlock(t *testing.T) {
	ws := &Workspace{ID: "ws-123"}
	app := newFakeCLI(ws)

	cmd := app.workspaceUnlockCommand()
	cmd.SetArgs([]string{"dev", "--organization", "acme-corp"})
	got := bytes.Buffer{}
	cmd.SetOut(&got)
	require.NoError(t, cmd.Execute())
	want := fmt.Sprintf("Successfully unlocked workspace %s\n", ws.Name)
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

type fakeCLIService struct {
	workspaces []*Workspace
	Service
}

func newFakeCLI(workspaces ...*Workspace) *CLI {
	return &CLI{Service: &fakeCLIService{workspaces: workspaces}}
}

func (f *fakeCLIService) GetWorkspace(context.Context, string) (*Workspace, error) {
	return f.workspaces[0], nil
}

func (f *fakeCLIService) GetWorkspaceByName(context.Context, string, string) (*Workspace, error) {
	return f.workspaces[0], nil
}

func (f *fakeCLIService) ListWorkspaces(ctx context.Context, opts ListOptions) (*resource.Page[*Workspace], error) {
	return resource.NewPage(f.workspaces, opts.PageOptions, nil), nil
}

func (f *fakeCLIService) UpdateWorkspace(ctx context.Context, workspaceID string, opts UpdateOptions) (*Workspace, error) {
	f.workspaces[0].Update(opts)
	return f.workspaces[0], nil
}

func (f *fakeCLIService) LockWorkspace(context.Context, string, *string) (*Workspace, error) {
	return f.workspaces[0], nil
}

func (f *fakeCLIService) UnlockWorkspace(context.Context, string, *string, bool) (*Workspace, error) {
	return f.workspaces[0], nil
}
