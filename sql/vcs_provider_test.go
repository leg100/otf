package sql

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVCSProvider_Create(t *testing.T) {
	ctx := context.Background()
	db := NewTestDB(t)
	org := CreateTestOrganization(t, db)
	provider := NewTestVCSProvider(t, org)

	defer db.DeleteVCSProvider(ctx, provider.Token())

	err := db.CreateVCSProvider(ctx, provider)
	require.NoError(t, err)
}

func TestVCSProvider_Get(t *testing.T) {
	ctx := context.Background()
	db := NewTestDB(t)
	org := CreateTestOrganization(t, db)
	want := CreateTestVCSProvider(t, db, org)

	got, err := db.GetVCSProvider(ctx, want.ID())
	require.NoError(t, err)

	assert.Equal(t, want, got)
}

func TestVCSProvider_List(t *testing.T) {
	ctx := context.Background()
	db := NewTestDB(t)
	org := CreateTestOrganization(t, db)
	provider1 := CreateTestVCSProvider(t, db, org)
	provider2 := CreateTestVCSProvider(t, db, org)
	provider3 := CreateTestVCSProvider(t, db, org)

	got, err := db.ListVCSProviders(ctx, org.Name())
	require.NoError(t, err)

	assert.Contains(t, got, provider1)
	assert.Contains(t, got, provider2)
	assert.Contains(t, got, provider3)
}

func TestVCSProvider_Delete(t *testing.T) {
	ctx := context.Background()
	db := NewTestDB(t)
	org := CreateTestOrganization(t, db)
	provider := CreateTestVCSProvider(t, db, org)

	err := db.DeleteVCSProvider(ctx, provider.ID())
	require.NoError(t, err)

	got, err := db.ListVCSProviders(ctx, org.Name())
	require.NoError(t, err)

	assert.Len(t, got, 0)
}
