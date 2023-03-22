package workspace

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"github.com/leg100/otf"
	"github.com/leg100/otf/auth"
	"github.com/leg100/otf/rbac"
	"github.com/stretchr/testify/require"
)

func NewTestService(t *testing.T, db otf.DB) *service {
	svc := &service{
		Logger: logr.Discard(),
		db:     &pgdb{db},
	}
	svc.Publisher = &otf.FakePublisher{}
	svc.organization = otf.NewAllowAllAuthorizer()
	svc.site = otf.NewAllowAllAuthorizer()
	svc.Authorizer = otf.NewAllowAllAuthorizer()
	return svc
}

func CreateTestWorkspace(t *testing.T, db otf.DB, organization string) *Workspace {
	ctx := context.Background()
	svc := NewTestService(t, db)

	createOptions := CreateOptions{
		Name:         otf.String(uuid.NewString()),
		Organization: &organization,
	}

	ws, err := svc.CreateWorkspace(ctx, createOptions)
	require.NoError(t, err)

	t.Cleanup(func() {
		svc.DeleteWorkspace(ctx, ws.ID)
	})
	return ws
}

func CreateTestWorkspacePermission(t *testing.T, db otf.DB, ws *Workspace, team *auth.Team, role rbac.Role) *Workspace {
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
	ws, err := NewWorkspace(CreateOptions{
		Name:         otf.String(uuid.NewString()),
		Organization: &organization,
	})
	require.NoError(t, err)
	return ws
}
