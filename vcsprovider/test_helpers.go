package vcsprovider

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/leg100/otf"
	"github.com/leg100/otf/inmem"
	"github.com/stretchr/testify/require"
)

func NewTestVCSProvider(t *testing.T, org *otf.Organization) *otf.VCSProvider {
	var organizationName string
	if org == nil {
		organizationName = uuid.NewString()
	} else {
		organizationName = org.Name
	}
	factory := &factory{inmem.NewTestCloudService()}
	provider, err := factory.new(createOptions{
		Organization: organizationName,
		// unit tests require a legitimate cloud name to avoid invalid foreign
		// key error upon insert/update
		Cloud: "github",
		Name:  uuid.NewString(),
		Token: uuid.NewString(),
	})
	require.NoError(t, err)
	return provider
}

func createTestVCSProvider(t *testing.T, db *pgdb, organization *otf.Organization) *otf.VCSProvider {
	provider := NewTestVCSProvider(t, organization)
	ctx := context.Background()

	err := db.create(ctx, provider)
	require.NoError(t, err)

	t.Cleanup(func() {
		db.delete(ctx, provider.ID)
	})
	return provider
}
