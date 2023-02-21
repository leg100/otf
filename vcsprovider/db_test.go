package vcsprovider

import (
	"context"
	"testing"

	"github.com/leg100/otf"
	"github.com/leg100/otf/inmem"
	"github.com/leg100/otf/organization"
	"github.com/leg100/otf/sql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDB(t *testing.T) {
	ctx := context.Background()
	db := newDB(sql.NewTestDB(t), inmem.NewTestCloudService())
	org := organization.CreateTestOrganization(t, db)

	t.Run("create", func(t *testing.T) {
		provider := NewTestVCSProvider(t, org)

		defer db.delete(ctx, provider.Token())

		err := db.create(ctx, provider)
		require.NoError(t, err)
	})

	t.Run("get", func(t *testing.T) {
		want := createTestVCSProvider(t, db, org)

		got, err := providerDB.get(ctx, want.ID())
		require.NoError(t, err)

		assert.Equal(t, want, got)
	})

	t.Run("list", func(t *testing.T) {
		ctx := context.Background()
		db := sql.NewTestDB(t)
		providerDB := newTestDB(t)
		org := sql.CreateTestOrganization(t, db)
		provider1 := createTestVCSProvider(t, providerDB, org)
		provider2 := createTestVCSProvider(t, providerDB, org)
		provider3 := createTestVCSProvider(t, providerDB, org)

		got, err := providerDB.list(ctx, org.Name())
		require.NoError(t, err)

		assert.Contains(t, got, provider1)
		assert.Contains(t, got, provider2)
		assert.Contains(t, got, provider3)
	})

	t.Run("delete", func(t *testing.T) {
		ctx := context.Background()
		db := sql.NewTestDB(t)
		providerDB := newTestDB(t)
		org := sql.CreateTestOrganization(t, db)
		provider := createTestVCSProvider(t, providerDB, org)

		err := providerDB.delete(ctx, provider.ID())
		require.NoError(t, err)

		got, err := providerDB.list(ctx, org.Name())
		require.NoError(t, err)

		assert.Len(t, got, 0)
	})
}

func createTestVCSProvider(t *testing.T, db *pgdb, organization otf.Organization) *VCSProvider {
	provider := NewTestVCSProvider(t, organization)
	ctx := context.Background()

	err := db.create(ctx, provider)
	require.NoError(t, err)

	t.Cleanup(func() {
		db.delete(ctx, provider.ID())
	})
	return provider
}
