package variable

import (
	"context"
	"testing"

	"github.com/leg100/otf"
	"github.com/leg100/otf/sql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVariable_Create(t *testing.T) {
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

	defer variableDB.DeleteVariable(ctx, v.ID())

	err := variableDB.CreateVariable(ctx, v)
	require.NoError(t, err)
}

func TestVariable_Update(t *testing.T) {
	ctx := context.Background()
	db := sql.NewTestDB(t)
	variableDB := &pgdb{db}
	org := sql.CreateTestOrganization(t, db)
	ws := sql.CreateTestWorkspace(t, db, org)
	country := createTestVariable(t, db, ws, otf.CreateVariableOptions{
		Key:      otf.String("country"),
		Value:    otf.String("belgium"),
		Category: VariableCategoryPtr(otf.CategoryTerraform),
	})

	got, err := variableDB.UpdateVariable(ctx, country.ID(), func(v *Variable) error {
		return v.Update(otf.UpdateVariableOptions{
			Value: otf.String("luxembourg"),
		})
	})
	require.NoError(t, err)

	assert.Equal(t, "luxembourg", got.Value())
}

func TestVariable_List(t *testing.T) {
	ctx := context.Background()
	db := sql.NewTestDB(t)
	variableDB := &pgdb{db}
	org := sql.CreateTestOrganization(t, db)
	ws := sql.CreateTestWorkspace(t, db, org)
	country := createTestVariable(t, db, ws, otf.CreateVariableOptions{
		Key:      otf.String("country"),
		Value:    otf.String("belgium"),
		Category: VariableCategoryPtr(otf.CategoryTerraform),
	})
	city := createTestVariable(t, db, ws, otf.CreateVariableOptions{
		Key:      otf.String("city"),
		Value:    otf.String("mechelen"),
		Category: VariableCategoryPtr(otf.CategoryTerraform),
	})

	got, err := variableDB.ListVariables(ctx, ws.ID())
	require.NoError(t, err)

	if assert.Equal(t, 2, len(got)) {
		assert.Contains(t, got, country)
		assert.Contains(t, got, city)
	}
}

func TestVariable_Get(t *testing.T) {
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

	got, err := variableDB.GetVariable(ctx, want.ID())
	require.NoError(t, err)

	assert.Equal(t, want, got)
}

func TestVariable_Delete(t *testing.T) {
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

	got, err := variableDB.DeleteVariable(ctx, want.ID())
	require.NoError(t, err)
	assert.Equal(t, want, got)
}

func createTestVariable(t *testing.T, db *pgdb, ws *otf.Workspace, opts otf.CreateVariableOptions) *Variable {
	ctx := context.Background()

	v, err := NewVariable(ws.ID(), opts)
	require.NoError(t, err)

	err = db.CreateVariable(ctx, v)
	require.NoError(t, err)

	t.Cleanup(func() {
		db.DeleteToken(ctx, v.ID())
	})
	return v
}
