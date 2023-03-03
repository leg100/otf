package testutil

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"github.com/leg100/otf"
	"github.com/leg100/otf/workspace"
	"github.com/stretchr/testify/require"
)

func NewWorkspaceService(t *testing.T, db otf.DB) *workspace.Service {
	svc, err := workspace.NewService(workspace.Options{
		Authorizer: NewAllowAllAuthorizer(),
		DB:         db,
		Logger:     logr.Discard(),
	})
	require.NoError(t, err)
	return svc
}

func NewWorkspace(t *testing.T, organization string, opts ...NewTestWorkspaceOption) *Workspace {
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

func CreateWorkspace(t *testing.T, db otf.DB, organization string, opts ...otf.NewTestWorkspaceOption) otf.Workspace {
	ctx := context.Background()
	ws := NewWorkspace(t, organization, opts...)
	err := workspaceService.CreateWorkspace(ctx, ws)
	require.NoError(t, err)

	t.Cleanup(func() {
		db.DeleteWorkspace(ctx, ws.ID)
	})
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
