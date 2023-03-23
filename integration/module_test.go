package integration

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/leg100/otf"
	"github.com/leg100/otf/module"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestModule(t *testing.T) {
	// perform all actions as superuser
	ctx := otf.AddSubjectToContext(context.Background(), &otf.Superuser{})

	t.Run("create", func(t *testing.T) {
		svc := setup(t, "")
		org := svc.createOrganization(t, ctx)

		_, err := svc.CreateModule(ctx, module.CreateOptions{
			Name:         uuid.NewString(),
			Provider:     uuid.NewString(),
			Organization: org.Name,
		})
		require.NoError(t, err)
	})

	t.Run("get", func(t *testing.T) {
		svc := setup(t, "")
		want := svc.createModule(t, ctx, nil)

		got, err := svc.GetModule(ctx, module.GetModuleOptions{
			Organization: want.Organization,
			Provider:     want.Provider,
			Name:         want.Name,
		})
		require.NoError(t, err)

		assert.Equal(t, want, got)
	})

	t.Run("get by id", func(t *testing.T) {
		svc := setup(t, "")
		want := svc.createModule(t, ctx, nil)

		got, err := svc.GetModuleByID(ctx, want.ID)
		require.NoError(t, err)

		assert.Equal(t, want, got)
	})

	t.Run("list", func(t *testing.T) {
		svc := setup(t, "")
		org := svc.createOrganization(t, ctx)
		module1 := svc.createModule(t, ctx, org)
		module2 := svc.createModule(t, ctx, org)
		module3 := svc.createModule(t, ctx, org)

		got, err := svc.ListModules(ctx, module.ListModulesOptions{
			Organization: org.Name,
		})
		require.NoError(t, err)

		assert.Contains(t, got, module1)
		assert.Contains(t, got, module2)
		assert.Contains(t, got, module3)
	})

	t.Run("delete", func(t *testing.T) {
		svc := setup(t, "")
		want := svc.createModule(t, ctx, nil)

		got, err := svc.DeleteModule(ctx, want.ID)
		require.NoError(t, err)

		assert.Equal(t, want, got)
	})
}
