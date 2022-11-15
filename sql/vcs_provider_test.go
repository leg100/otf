package sql

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVCSProvider_Create(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	org := createTestOrganization(t, db)
	provider := newTestVCSProvider(org)

	defer db.DeleteVCSProvider(ctx, provider.Token())

	err := db.CreateVCSProvider(ctx, provider)
	require.NoError(t, err)
}

func TestVCSProvider_Get(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	org := createTestOrganization(t, db)
	want := createTestVCSProvider(t, db, org)

	got, err := db.GetVCSProvider(ctx, want.ID())
	require.NoError(t, err)

	assert.Equal(t, want, got)
}

func TestVCSProvider_List(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	org := createTestOrganization(t, db)
	provider1 := createTestVCSProvider(t, db, org)
	provider2 := createTestVCSProvider(t, db, org)
	provider3 := createTestVCSProvider(t, db, org)

	got, err := db.ListVCSProviders(ctx, org.Name())
	require.NoError(t, err)

	assert.Contains(t, got, provider1)
	assert.Contains(t, got, provider2)
	assert.Contains(t, got, provider3)
}

func TestVCSProvider_Delete(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	org := createTestOrganization(t, db)
	provider := createTestVCSProvider(t, db, org)

	err := db.DeleteVCSProvider(ctx, provider.ID())
	require.NoError(t, err)

	got, err := db.ListVCSProviders(ctx, org.Name())
	require.NoError(t, err)

	assert.Len(t, got, 0)
}
