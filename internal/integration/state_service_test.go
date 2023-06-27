package integration

import (
	"os"
	"testing"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/state"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegration_StateService(t *testing.T) {
	integrationTest(t)

	t.Run("create", func(t *testing.T) {
		svc, _, ctx := setup(t, nil)
		ws := svc.createWorkspace(t, ctx, nil)
		file, err := os.ReadFile("./testdata/terraform.tfstate")
		require.NoError(t, err)

		_, err = svc.CreateStateVersion(ctx, state.CreateStateVersionOptions{
			State:       file,
			WorkspaceID: internal.String(ws.ID),
		})
		require.NoError(t, err)
	})

	t.Run("get", func(t *testing.T) {
		svc, _, ctx := setup(t, nil)
		want := svc.createStateVersion(t, ctx, nil)

		got, err := svc.GetStateVersion(ctx, want.ID)
		require.NoError(t, err)

		assert.Equal(t, want, got)
	})

	t.Run("get not found error", func(t *testing.T) {
		svc, _, ctx := setup(t, nil)

		_, err := svc.GetStateVersion(ctx, "sv-99999")
		require.Equal(t, internal.ErrResourceNotFound, err)
	})

	// Get current creates two state versions and checks the second one is made
	// the current state version for a workspace.
	t.Run("get current", func(t *testing.T) {
		svc, _, ctx := setup(t, nil)
		ws := svc.createWorkspace(t, ctx, nil)
		_ = svc.createStateVersion(t, ctx, ws)
		want := svc.createStateVersion(t, ctx, ws)

		got, err := svc.GetCurrentStateVersion(ctx, want.WorkspaceID)
		require.NoError(t, err)

		assert.Equal(t, want, got)
	})

	t.Run("get current not found error", func(t *testing.T) {
		svc, _, ctx := setup(t, nil)

		_, err := svc.GetCurrentStateVersion(ctx, "ws-99999")
		assert.Equal(t, internal.ErrResourceNotFound, err)
	})

	t.Run("list", func(t *testing.T) {
		svc, _, ctx := setup(t, nil)
		ws := svc.createWorkspace(t, ctx, nil)
		sv1 := svc.createStateVersion(t, ctx, ws)
		sv2 := svc.createStateVersion(t, ctx, ws)

		got, err := svc.ListStateVersions(ctx, ws.ID, internal.ListOptions{})
		require.NoError(t, err)
		assert.Contains(t, got.Items, sv1)
		assert.Contains(t, got.Items, sv2)
	})

	// Listing state versions for a non-existent workspace should produce an
	// error
	t.Run("list not found error", func(t *testing.T) {
		svc, _, ctx := setup(t, nil)

		_, err := svc.ListStateVersions(ctx, "ws-does-not-exist", internal.ListOptions{})
		assert.Equal(t, internal.ErrResourceNotFound, err)
	})

	t.Run("delete", func(t *testing.T) {
		svc, _, ctx := setup(t, nil)
		ws := svc.createWorkspace(t, ctx, nil)
		want := svc.createStateVersion(t, ctx, ws)
		current := svc.createStateVersion(t, ctx, ws)

		err := svc.DeleteStateVersion(ctx, want.ID)
		require.NoError(t, err)

		_, err = svc.GetStateVersion(ctx, want.ID)
		assert.Equal(t, internal.ErrResourceNotFound, err)

		t.Run("deleting current version not allowed", func(t *testing.T) {
			err := svc.DeleteStateVersion(ctx, current.ID)
			assert.Equal(t, state.ErrCurrentVersionDeletionAttempt, err)
		})
	})
}
