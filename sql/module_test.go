package sql

import (
	"context"
	"testing"

	"github.com/leg100/otf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestModule_Create(t *testing.T) {
	ctx := context.Background()
	db := NewTestDB(t)
	org := CreateTestOrganization(t, db)
	module := otf.NewTestModule(org)

	defer db.DeleteModule(ctx, module.ID())

	err := db.CreateModule(ctx, module)
	require.NoError(t, err)
}

func TestModule_Get(t *testing.T) {
	ctx := context.Background()
	db := NewTestDB(t)
	org := CreateTestOrganization(t, db)
	want := CreateTestModule(t, db, org)

	got, err := db.GetModule(ctx, otf.GetModuleOptions{
		Organization: org.Name(),
		Provider:     want.Provider(),
		Name:         want.Name(),
	})
	require.NoError(t, err)

	assert.Equal(t, want, got)
}

func TestModule_GetByID(t *testing.T) {
	ctx := context.Background()
	db := NewTestDB(t)
	org := CreateTestOrganization(t, db)
	want := CreateTestModule(t, db, org)

	got, err := db.GetModuleByID(ctx, want.ID())
	require.NoError(t, err)

	assert.Equal(t, want, got)
}

func TestModule_List(t *testing.T) {
	ctx := context.Background()
	db := NewTestDB(t)
	org := CreateTestOrganization(t, db)
	module1 := CreateTestModule(t, db, org)
	module2 := CreateTestModule(t, db, org)
	module3 := CreateTestModule(t, db, org)

	got, err := db.ListModules(ctx, otf.ListModulesOptions{
		Organization: org.Name(),
	})
	require.NoError(t, err)

	assert.Contains(t, got, module1)
	assert.Contains(t, got, module2)
	assert.Contains(t, got, module3)
}

func TestModule_Delete(t *testing.T) {
	ctx := context.Background()
	db := NewTestDB(t)
	org := CreateTestOrganization(t, db)
	module := CreateTestModule(t, db, org)

	err := db.DeleteModule(ctx, module.ID())
	require.NoError(t, err)

	got, err := db.ListModules(ctx, otf.ListModulesOptions{Organization: org.Name()})
	require.NoError(t, err)

	assert.Len(t, got, 0)
}
