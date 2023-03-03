package variable

import (
	"context"
	"testing"

	"github.com/leg100/otf"
	"github.com/leg100/otf/sql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDB_Create(t *testing.T) {
	ctx := context.Background()
	db := sql.NewTestDB(t)
	variableDB := &pgdb{db}
	org := sql.CreateTestOrganization(t, db)
	ws := sql.CreateTestWorkspace(t, db, org)
	v := NewTestVariable(t, ws, otf.CreateVariableOptions{
		Key:      otf.String("foo"),
		Value:    otf.String("bar"),
		Category: VariableCategoryPtr(otf.CategoryTerraform),
	})

	defer variableDB.delete(ctx, v.ID)

	err := variableDB.create(ctx, v)
	require.NoError(t, err)
}

func TestDB_Update(t *testing.T) {
	ctx := context.Background()
	db := sql.NewTestDB(t)
	variableDB := &pgdb{db}
	org := sql.CreateTestOrganization(t, db)
	ws := sql.CreateTestWorkspace(t, db, org)
	country := createTestVariable(t, variableDB, ws, otf.CreateVariableOptions{
		Key:      otf.String("country"),
		Value:    otf.String("belgium"),
		Category: VariableCategoryPtr(otf.CategoryTerraform),
	})

	got, err := variableDB.update(ctx, country.ID, func(v *Variable) error {
		return v.Update(otf.UpdateVariableOptions{
			Value: otf.String("luxembourg"),
		})
	})
	require.NoError(t, err)

	assert.Equal(t, "luxembourg", got.Value())
}

func TestDB_List(t *testing.T) {
	ctx := context.Background()
	db := sql.NewTestDB(t)
	variableDB := &pgdb{db}
	org := sql.CreateTestOrganization(t, db)
	ws := sql.CreateTestWorkspace(t, db, org)
	country := createTestVariable(t, variableDB, ws, otf.CreateVariableOptions{
		Key:      otf.String("country"),
		Value:    otf.String("belgium"),
		Category: VariableCategoryPtr(otf.CategoryTerraform),
	})
	city := createTestVariable(t, variableDB, ws, otf.CreateVariableOptions{
		Key:      otf.String("city"),
		Value:    otf.String("mechelen"),
		Category: VariableCategoryPtr(otf.CategoryTerraform),
	})

	got, err := variableDB.list(ctx, ws.ID)
	require.NoError(t, err)

	if assert.Equal(t, 2, len(got)) {
		assert.Contains(t, got, country)
		assert.Contains(t, got, city)
	}
}

func TestDB_Get(t *testing.T) {
	ctx := context.Background()
	db := sql.NewTestDB(t)
	variableDB := &pgdb{db}
	org := sql.CreateTestOrganization(t, db)
	ws := sql.CreateTestWorkspace(t, db, org)
	want := createTestVariable(t, variableDB, ws, otf.CreateVariableOptions{
		Key:      otf.String("country"),
		Value:    otf.String("belgium"),
		Category: VariableCategoryPtr(otf.CategoryTerraform),
	})

	got, err := variableDB.get(ctx, want.ID)
	require.NoError(t, err)

	assert.Equal(t, want, got)
}

func TestDB_Delete(t *testing.T) {
	ctx := context.Background()
	db := sql.NewTestDB(t)
	variableDB := &pgdb{db}
	org := sql.CreateTestOrganization(t, db)
	ws := sql.CreateTestWorkspace(t, db, org)
	want := createTestVariable(t, variableDB, ws, otf.CreateVariableOptions{
		Key:      otf.String("country"),
		Value:    otf.String("belgium"),
		Category: VariableCategoryPtr(otf.CategoryTerraform),
	})

	got, err := variableDB.delete(ctx, want.ID)
	require.NoError(t, err)
	assert.Equal(t, want, got)
}

func createTestVariable(t *testing.T, db *pgdb, ws *otf.Workspace, opts otf.CreateVariableOptions) *Variable {
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
