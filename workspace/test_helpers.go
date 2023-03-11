package workspace

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"github.com/leg100/otf"
	"github.com/stretchr/testify/require"
)

func NewTestService(t *testing.T, db otf.DB) *Service {
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
