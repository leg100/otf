package integration

import (
	"testing"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/variable"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVariable(t *testing.T) {
	integrationTest(t)

	t.Run("create", func(t *testing.T) {
		svc, _, ctx := setup(t)
		ws := svc.createWorkspace(t, ctx, nil)

		_, err := svc.Variables.CreateWorkspaceVariable(ctx, ws.ID, variable.CreateVariableOptions{
			Key:      new("foo"),
			Value:    new("bar"),
			Category: internal.Ptr(variable.CategoryTerraform),
		})
		require.NoError(t, err)
	})

	t.Run("update", func(t *testing.T) {
		svc, _, ctx := setup(t)
		v := svc.createVariable(t, ctx, nil, nil)

		got, err := svc.Variables.UpdateWorkspaceVariable(ctx, v.ID, variable.UpdateVariableOptions{
			Value: new("luxembourg"),
		})
		require.NoError(t, err)

		assert.Equal(t, "luxembourg", got.Variable.Value)
	})

	t.Run("list", func(t *testing.T) {
		svc, _, ctx := setup(t)
		ws := svc.createWorkspace(t, ctx, nil)
		v1 := svc.createVariable(t, ctx, ws, nil)
		v2 := svc.createVariable(t, ctx, ws, nil)

		got, err := svc.Variables.ListWorkspaceVariables(ctx, ws.ID)
		require.NoError(t, err)

		if assert.Equal(t, 2, len(got)) {
			assert.Contains(t, got, v1)
			assert.Contains(t, got, v2)
		}
	})

	t.Run("get", func(t *testing.T) {
		svc, _, ctx := setup(t)
		want := svc.createVariable(t, ctx, nil, nil)

		got, err := svc.Variables.GetWorkspaceVariable(ctx, want.ID)
		require.NoError(t, err)

		assert.Equal(t, want, got.Variable)
	})

	t.Run("delete", func(t *testing.T) {
		svc, _, ctx := setup(t)
		want := svc.createVariable(t, ctx, nil, nil)

		got, err := svc.Variables.DeleteWorkspaceVariable(ctx, want.ID)
		require.NoError(t, err)
		assert.Equal(t, want, got.Variable)
	})
}
