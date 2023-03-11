package workspace

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"github.com/leg100/otf"
	"github.com/leg100/otf/rbac"
	"github.com/stretchr/testify/require"
)

func NewTestService(t *testing.T, db otf.DB) *service {
	service := NewService(Options{
		DB:        db,
		Logger:    logr.Discard(),
		Publisher: &otf.FakePublisher{},
	})
	service.organization = otf.NewAllowAllAuthorizer()
	service.site = otf.NewAllowAllAuthorizer()
	service.Authorizer = otf.NewAllowAllAuthorizer()
	return service
}

func CreateTestWorkspace(t *testing.T, db otf.DB, organization string) *Workspace {
	ctx := context.Background()
	svc := NewTestService(t, db)
	ws, err := svc.CreateWorkspace(ctx, CreateWorkspaceOptions{
		Name:         otf.String(uuid.NewString()),
		Organization: &organization,
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		svc.DeleteWorkspace(ctx, ws.ID)
	})
	return ws
}

func CreateTestWorkspacePermission(t *testing.T, db otf.DB, ws *Workspace, team *otf.Team, role rbac.Role) *Workspace {
	ctx := context.Background()
	workspaceDB := &pgdb{db}
	err := workspaceDB.SetWorkspacePermission(ctx, ws.ID, team.Name, role)
	require.NoError(t, err)

	t.Cleanup(func() {
		workspaceDB.UnsetWorkspacePermission(ctx, ws.ID, team.Name)
	})
	return ws
}

func newTestWorkspace(t *testing.T, organization string) *Workspace {
	ws, err := NewWorkspace(CreateWorkspaceOptions{
		Name:         otf.String(uuid.NewString()),
		Organization: &organization,
	})
	require.NoError(t, err)
	return ws
}
