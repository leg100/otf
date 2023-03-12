package integration

import (
	"context"
	"testing"

	"github.com/leg100/otf"
	"github.com/leg100/otf/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepo(t *testing.T) {
	t.Run("create workspace connection", func(t *testing.T) {
		ws := workspace.CreateTestWorkspace(t, db, org.Name)
		hook := createTestHook(t, db, nil)

		err := db.createConnection(ctx, hook.id, otf.ConnectOptions{
			ConnectionType: otf.WorkspaceConnection,
			VCSProviderID:  provider.ID,
			ResourceID:     ws.ID,
			RepoPath:       hook.identifier,
		})
		assert.NoError(t, err)
	})

	t.Run("create module connection", func(t *testing.T) {
		module := createTestModule(t)
		hook := createTestHook(t, db, nil)

		err := db.createConnection(ctx, hook.id, otf.ConnectOptions{
			ConnectionType: otf.ModuleConnection,
			VCSProviderID:  provider.ID,
			ResourceID:     module.ID,
			RepoPath:       hook.identifier,
		})
		assert.NoError(t, err)
	})

	t.Run("delete workspace connection", func(t *testing.T) {
		ws := workspace.CreateTestWorkspace(t, db, org.Name)
		hook := createTestHook(t, db, nil)
		_ = createTestConnection(t, db, provider, hook, otf.WorkspaceConnection, ws.ID)

		gotHookID, gotProviderID, err := db.deleteConnection(ctx, otf.DisconnectOptions{
			ConnectionType: otf.WorkspaceConnection,
			ResourceID:     ws.ID,
		})
		assert.NoError(t, err)
		assert.Equal(t, hook.id, gotHookID)
		assert.Equal(t, provider.ID, gotProviderID)
	})

	t.Run("delete module connection", func(t *testing.T) {
		module := createTestModule(t)
		hook := createTestHook(t, db, nil)
		_ = createTestConnection(t, db, provider, hook, otf.ModuleConnection, module.ID)

		gotHookID, gotProviderID, err := db.deleteConnection(ctx, otf.DisconnectOptions{
			ConnectionType: otf.ModuleConnection,
			ResourceID:     module.ID,
		})
		assert.NoError(t, err)
		assert.Equal(t, hook.id, gotHookID)
		assert.Equal(t, provider.ID, gotProviderID)
	})

	t.Run("count connections", func(t *testing.T) {
		ws1 := workspace.CreateTestWorkspace(t, db, org.Name)
		ws2 := workspace.CreateTestWorkspace(t, db, org.Name)
		module1 := createTestModule(t)

		hook := createTestHook(t, db, nil)
		_ = createTestConnection(t, db, provider, hook, otf.WorkspaceConnection, ws1.ID)
		_ = createTestConnection(t, db, provider, hook, otf.WorkspaceConnection, ws2.ID)
		_ = createTestConnection(t, db, provider, hook, otf.ModuleConnection, module1.ID)

		got, err := db.countConnections(ctx, hook.id)
		require.NoError(t, err)
		assert.Equal(t, 3, *got)
	})
}

func createTestConnection(t *testing.T, db *pgdb, provider *otf.VCSProvider, hook *hook, connType otf.ConnectionType, resourceID string) *otf.Connection {
	ctx := context.Background()

	err := db.createConnection(ctx, hook.id, otf.ConnectOptions{
		ConnectionType: connType,
		VCSProviderID:  provider.ID,
		ResourceID:     resourceID,
		RepoPath:       hook.identifier,
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		db.deleteConnection(ctx, otf.DisconnectOptions{
			ConnectionType: connType,
			ResourceID:     resourceID,
		})
	})
	return &otf.Connection{
		VCSProviderID: provider.ID,
		Repo:          hook.id,
	}
}
