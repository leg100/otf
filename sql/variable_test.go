package sql

import (
	"context"
	"testing"

	"github.com/leg100/otf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVariable_Create(t *testing.T) {
	ctx := context.Background()
	db := NewTestDB(t)
	org := CreateTestOrganization(t, db)
	ws := CreateTestWorkspace(t, db, org)
	v := otf.NewTestVariable(t, ws, otf.CreateVariableOptions{
		Key:      otf.String("foo"),
		Value:    otf.String("bar"),
		Category: otf.VariableCategoryPtr(otf.CategoryTerraform),
	})

	defer db.DeleteVariable(ctx, v.ID())

	err := db.CreateVariable(ctx, v)
	require.NoError(t, err)
}

func TestVariable_Update(t *testing.T) {
	ctx := context.Background()
	db := NewTestDB(t)
	org := CreateTestOrganization(t, db)
	ws := CreateTestWorkspace(t, db, org)
	country := createTestVariable(t, db, ws, otf.CreateVariableOptions{
		Key:      otf.String("country"),
		Value:    otf.String("belgium"),
		Category: otf.VariableCategoryPtr(otf.CategoryTerraform),
	})

	got, err := db.UpdateVariable(ctx, country.ID(), func(v *otf.Variable) error {
		return v.Update(otf.UpdateVariableOptions{
			Value: otf.String("luxembourg"),
		})
	})
	require.NoError(t, err)

	assert.Equal(t, "luxembourg", got.Value())
}

func TestVariable_List(t *testing.T) {
	ctx := context.Background()
	db := NewTestDB(t)
	org := CreateTestOrganization(t, db)
	ws := CreateTestWorkspace(t, db, org)
	country := createTestVariable(t, db, ws, otf.CreateVariableOptions{
		Key:      otf.String("country"),
		Value:    otf.String("belgium"),
		Category: otf.VariableCategoryPtr(otf.CategoryTerraform),
	})
	city := createTestVariable(t, db, ws, otf.CreateVariableOptions{
		Key:      otf.String("city"),
		Value:    otf.String("mechelen"),
		Category: otf.VariableCategoryPtr(otf.CategoryTerraform),
	})

	got, err := db.ListVariables(ctx, ws.ID())
	require.NoError(t, err)

	if assert.Equal(t, 2, len(got)) {
		assert.Contains(t, got, country)
		assert.Contains(t, got, city)
	}
}

func TestVariable_Get(t *testing.T) {
	ctx := context.Background()
	db := NewTestDB(t)
	org := CreateTestOrganization(t, db)
	ws := CreateTestWorkspace(t, db, org)
	want := createTestVariable(t, db, ws, otf.CreateVariableOptions{
		Key:      otf.String("country"),
		Value:    otf.String("belgium"),
		Category: otf.VariableCategoryPtr(otf.CategoryTerraform),
	})

	got, err := db.GetVariable(ctx, want.ID())
	require.NoError(t, err)

	assert.Equal(t, want, got)
}

func TestVariable_Delete(t *testing.T) {
	ctx := context.Background()
	db := NewTestDB(t)
	org := CreateTestOrganization(t, db)
	ws := CreateTestWorkspace(t, db, org)
	want := createTestVariable(t, db, ws, otf.CreateVariableOptions{
		Key:      otf.String("country"),
		Value:    otf.String("belgium"),
		Category: otf.VariableCategoryPtr(otf.CategoryTerraform),
	})

	got, err := db.DeleteVariable(ctx, want.ID())
	require.NoError(t, err)
	assert.Equal(t, want, got)
}
