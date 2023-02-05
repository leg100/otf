package vcsprovider

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/leg100/otf"
	"github.com/leg100/otf/inmem"
	"github.com/leg100/otf/sql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDB_Create(t *testing.T) {
	ctx := context.Background()
	db := sql.NewTestDB(t)
	providerDB := newTestDB(t)
	org := sql.CreateTestOrganization(t, db)
	provider := newTestVCSProvider(t, org)

	defer providerDB.delete(ctx, provider.Token())

	err := providerDB.create(ctx, provider)
	require.NoError(t, err)
}

func TestDB_Get(t *testing.T) {
	ctx := context.Background()
	db := sql.NewTestDB(t)
	providerDB := newTestDB(t)
	org := sql.CreateTestOrganization(t, db)
	want := createTestVCSProvider(t, providerDB, org)

	got, err := providerDB.get(ctx, want.ID())
	require.NoError(t, err)

	assert.Equal(t, want, got)
}

func TestDB_List(t *testing.T) {
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
}

func TestDB_Delete(t *testing.T) {
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
}

func newTestVCSProvider(t *testing.T, org *otf.Organization) *VCSProvider {
	factory := &factory{inmem.NewTestCloudService()}
	provider, err := factory.new(createOptions{
		Organization: org.Name(),
		// unit tests require a legitimate cloud name to avoid invalid foreign
		// key error upon insert/update
		Cloud: "github",
		Name:  uuid.NewString(),
		Token: uuid.NewString(),
	})
	require.NoError(t, err)
	return provider
}

func createTestVCSProvider(t *testing.T, db *pgdb, organization *otf.Organization) *VCSProvider {
	provider := newTestVCSProvider(t, organization)
	ctx := context.Background()

	err := db.create(ctx, provider)
	require.NoError(t, err)

	t.Cleanup(func() {
		db.delete(ctx, provider.ID())
	})
	return provider
}

func newTestDB(t *testing.T) *pgdb {
	return &pgdb{
		Database: sql.NewTestDB(t),
		factory: factory{
			Service: &factory{inmem.NewTestCloudService()},
		},
	}
}
