package module

import (
	"context"
	"testing"

	"github.com/leg100/otf/organization"
	"github.com/leg100/otf/sql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDB(t *testing.T) {
	ctx := context.Background()
	db := &pgdb{sql.NewTestDB(t)}

	t.Run("create", func(t *testing.T) {
		org := organization.CreateTestOrganization(t, db)
		module := NewTestModule(org)

		defer db.delete(ctx, module.ID)

		err := db.createModule(ctx, module)
		require.NoError(t, err)
	})

	t.Run("get", func(t *testing.T) {
		org := organization.CreateTestOrganization(t, db)
		want := createTestModule(t, db, org)

		got, err := db.getModule(ctx, GetModuleOptions{
			Organization: org.Name,
			Provider:     want.Provider,
			Name:         want.Name,
		})
		require.NoError(t, err)

		assert.Equal(t, want, got)
	})

	t.Run("get by id", func(t *testing.T) {
		org := organization.CreateTestOrganization(t, db)
		want := createTestModule(t, db, org)

		got, err := db.getModuleByID(ctx, want.ID)
		require.NoError(t, err)

		assert.Equal(t, want, got)
	})

	t.Run("list", func(t *testing.T) {
		org := organization.CreateTestOrganization(t, db)
		module1 := createTestModule(t, db, org)
		module2 := createTestModule(t, db, org)
		module3 := createTestModule(t, db, org)

		got, err := db.listModules(ctx, ListModulesOptions{
			Organization: org.Name,
		})
		require.NoError(t, err)

		assert.Contains(t, got, module1)
		assert.Contains(t, got, module2)
		assert.Contains(t, got, module3)
	})

	t.Run("delete", func(t *testing.T) {
		org := organization.CreateTestOrganization(t, db)
		module := createTestModule(t, db, org)

		err := db.delete(ctx, module.ID)
		require.NoError(t, err)

		got, err := db.listModules(ctx, ListModulesOptions{Organization: org.Name})
		require.NoError(t, err)

		assert.Len(t, got, 0)
	})
}
