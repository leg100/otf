package integration

import (
	"context"
	"testing"

	"github.com/leg100/otf"
	"github.com/leg100/otf/variable"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVariable(t *testing.T) {
	// perform all actions as superuser
	ctx := otf.AddSubjectToContext(context.Background(), &otf.Superuser{})

	t.Run("create", func(t *testing.T) {
		svc := setup(t, nil)
		ws := svc.createWorkspace(t, ctx, nil)

		_, err := svc.CreateVariable(ctx, ws.ID, variable.CreateVariableOptions{
			Key:      otf.String("foo"),
			Value:    otf.String("bar"),
			Category: variable.VariableCategoryPtr(variable.CategoryTerraform),
		})
		require.NoError(t, err)
	})

	t.Run("update", func(t *testing.T) {
		svc := setup(t, nil)
		v := svc.createVariable(t, ctx, nil)

		got, err := svc.UpdateVariable(ctx, v.ID, variable.UpdateVariableOptions{
			Value: otf.String("luxembourg"),
		})
		require.NoError(t, err)

		assert.Equal(t, "luxembourg", got.Value)
	})

	t.Run("list", func(t *testing.T) {
		svc := setup(t, nil)
		ws := svc.createWorkspace(t, ctx, nil)
		v1 := svc.createVariable(t, ctx, ws)
		v2 := svc.createVariable(t, ctx, ws)

		got, err := svc.ListVariables(ctx, ws.ID)
		require.NoError(t, err)

		if assert.Equal(t, 2, len(got)) {
			assert.Contains(t, got, v1)
			assert.Contains(t, got, v2)
		}
	})

	t.Run("get", func(t *testing.T) {
		svc := setup(t, nil)
		want := svc.createVariable(t, ctx, nil)

		got, err := svc.GetVariable(ctx, want.ID)
		require.NoError(t, err)

		assert.Equal(t, want, got)
	})

	t.Run("delete", func(t *testing.T) {
		svc := setup(t, nil)
		want := svc.createVariable(t, ctx, nil)

		got, err := svc.DeleteVariable(ctx, want.ID)
		require.NoError(t, err)
		assert.Equal(t, want, got)
	})
}
