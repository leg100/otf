package testutil

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func CreateTestWorkspace(t *testing.T, db otf.DB, organization string, opts ...otf.NewTestWorkspaceOption) otf.Workspace {
	ctx := context.Background()
	ws := NewTestWorkspace(t, organization, opts...)
	wsdb := newPGDB(db)
	err := wsdb.CreateWorkspace(ctx, ws)
	require.NoError(t, err)

	t.Cleanup(func() {
		db.DeleteWorkspace(ctx, ws.ID())
	})
	return ws
}

func NewTestWorkspace(t *testing.T, organization string, opts ...NewTestWorkspaceOption) *Workspace {
	createOpts := CreateWorkspaceOptions{
		Name:         otf.String(uuid.NewString()),
		Organization: otf.String(organization),
	}
	ws, err := NewWorkspace(createOpts)
	require.NoError(t, err)
	for _, fn := range opts {
		fn(ws)
	}
	return ws
}

type NewTestWorkspaceOption func(*Workspace)

func AutoApply() NewTestWorkspaceOption {
	return func(ws *Workspace) {
		ws.autoApply = true
	}
}

func WithRepo(repo *WorkspaceRepo) NewTestWorkspaceOption {
	return func(ws *Workspace) {
		ws.repo = repo
	}
}

func WorkingDirectory(relativePath string) NewTestWorkspaceOption {
	return func(ws *Workspace) {
		ws.workingDirectory = relativePath
	}
}
