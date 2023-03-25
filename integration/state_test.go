package integration

import (
	"context"
	"os"
	"testing"

	"github.com/leg100/otf"
	"github.com/leg100/otf/state"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestState(t *testing.T) {
	// perform all actions as superuser
	ctx := otf.AddSubjectToContext(context.Background(), &otf.Superuser{})

	t.Run("create", func(t *testing.T) {
		svc := setup(t, nil)
		ws := svc.createWorkspace(t, ctx, nil)
		file, err := os.ReadFile("./testdata/terraform.tfstate")
		require.NoError(t, err)

		_, err = svc.CreateStateVersion(ctx, state.CreateStateVersionOptions{
			State:       file,
			WorkspaceID: otf.String(ws.ID),
		})
		require.NoError(t, err)
	})

	t.Run("get", func(t *testing.T) {
		svc := setup(t, nil)
		want := svc.createStateVersion(t, ctx, nil)

		got, err := svc.GetStateVersion(ctx, want.ID)
		require.NoError(t, err)

		assert.Equal(t, want, got)
	})

	t.Run("get not found error", func(t *testing.T) {
		svc := setup(t, nil)

		_, err := svc.GetStateVersion(ctx, "sv-99999")
		require.Equal(t, otf.ErrResourceNotFound, err)
	})

	t.Run("get current", func(t *testing.T) {
		svc := setup(t, nil)
		ws := svc.createWorkspace(t, ctx, nil)
		_ = svc.createStateVersion(t, ctx, ws)
		want := svc.createStateVersion(t, ctx, ws)

		got, err := svc.GetCurrentStateVersion(ctx, want.WorkspaceID)
		require.NoError(t, err)

		assert.Equal(t, want, got)
	})

	t.Run("get current not found error", func(t *testing.T) {
		svc := setup(t, nil)

		_, err := svc.GetCurrentStateVersion(ctx, "ws-99999")
		assert.Equal(t, otf.ErrResourceNotFound, err)
	})

	t.Run("list", func(t *testing.T) {
		svc := setup(t, nil)
		ws := svc.createWorkspace(t, ctx, nil)
		sv1 := svc.createStateVersion(t, ctx, ws)
		sv2 := svc.createStateVersion(t, ctx, ws)

		got, err := svc.ListStateVersions(ctx, state.StateVersionListOptions{
			Workspace:    ws.Name,
			Organization: ws.Organization,
		})
		require.NoError(t, err)
		assert.Contains(t, got.Items, sv1)
		assert.Contains(t, got.Items, sv2)
	})

	t.Run("list not found error", func(t *testing.T) {
		svc := setup(t, nil)

		_, err := svc.ListStateVersions(ctx, state.StateVersionListOptions{
			Workspace:    "ws-999",
			Organization: "acme-corp",
		})
		assert.Equal(t, otf.ErrResourceNotFound, err)
	})
}
