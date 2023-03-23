package variable

import (
	"context"
	"testing"

	"github.com/leg100/otf"
	"github.com/leg100/otf/organization"
	"github.com/leg100/otf/sql"
	"github.com/leg100/otf/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDB_Create(t *testing.T) {
	ctx := context.Background()
	db := &pgdb{sql.NewTestDB(t)}

	t.Run("create", func(t *testing.T) {
		org := organization.CreateTestOrganization(t, db)
		ws := workspace.CreateTestWorkspace(t, db, org.Name)
		v := NewTestVariable(t, ws.ID, CreateVariableOptions{
			Key:      otf.String("foo"),
			Value:    otf.String("bar"),
			Category: VariableCategoryPtr(CategoryTerraform),
		})

		defer db.delete(ctx, v.ID)

		err := db.create(ctx, v)
		require.NoError(t, err)
	})

	t.Run("update", func(t *testing.T) {
		org := organization.CreateTestOrganization(t, db)
		ws := workspace.CreateTestWorkspace(t, db, org.Name)
		country := createTestVariable(t, db, ws, CreateVariableOptions{
			Key:      otf.String("country"),
			Value:    otf.String("belgium"),
			Category: VariableCategoryPtr(CategoryTerraform),
		})

		got, err := db.update(ctx, country.ID, func(v *Variable) error {
			return v.Update(UpdateVariableOptions{
				Value: otf.String("luxembourg"),
			})
		})
		require.NoError(t, err)

		assert.Equal(t, "luxembourg", got.Value)
	})

	t.Run("list", func(t *testing.T) {
		org := organization.CreateTestOrganization(t, db)
		ws := workspace.CreateTestWorkspace(t, db, org.Name)
		country := createTestVariable(t, db, ws, CreateVariableOptions{
			Key:      otf.String("country"),
			Value:    otf.String("belgium"),
			Category: VariableCategoryPtr(CategoryTerraform),
		})
		city := createTestVariable(t, db, ws, CreateVariableOptions{
			Key:      otf.String("city"),
			Value:    otf.String("mechelen"),
			Category: VariableCategoryPtr(CategoryTerraform),
		})

		got, err := db.list(ctx, ws.ID)
		require.NoError(t, err)

		if assert.Equal(t, 2, len(got)) {
			assert.Contains(t, got, country)
			assert.Contains(t, got, city)
		}
	})

	t.Run("get", func(t *testing.T) {
		org := organization.CreateTestOrganization(t, db)
		ws := workspace.CreateTestWorkspace(t, db, org.Name)
		want := createTestVariable(t, db, ws, CreateVariableOptions{
			Key:      otf.String("country"),
			Value:    otf.String("belgium"),
			Category: VariableCategoryPtr(CategoryTerraform),
		})

		got, err := db.get(ctx, want.ID)
		require.NoError(t, err)

		assert.Equal(t, want, got)
	})

	t.Run("delete", func(t *testing.T) {
		org := organization.CreateTestOrganization(t, db)
		ws := workspace.CreateTestWorkspace(t, db, org.Name)
		want := createTestVariable(t, db, ws, CreateVariableOptions{
			Key:      otf.String("country"),
			Value:    otf.String("belgium"),
			Category: VariableCategoryPtr(CategoryTerraform),
		})

		got, err := db.delete(ctx, want.ID)
		require.NoError(t, err)
		assert.Equal(t, want, got)
	})
}

func createTestVariable(t *testing.T, db *pgdb, ws *workspace.Workspace, opts CreateVariableOptions) *Variable {
	ctx := context.Background()

	v, err := NewVariable(ws.ID, opts)
	require.NoError(t, err)

	err = db.create(ctx, v)
	require.NoError(t, err)

	t.Cleanup(func() {
		db.delete(ctx, v.ID)
	})
	return v
}
