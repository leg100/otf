package vcsprovider

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/leg100/otf/inmem"
	"github.com/leg100/otf/organization"
	"github.com/stretchr/testify/require"
)

func newTestVCSProvider(t *testing.T, org *organization.Organization) *VCSProvider {
	var organizationName string
	if org == nil {
		organizationName = uuid.NewString()
	} else {
		organizationName = org.Name
	}
	factory := &factory{inmem.NewTestCloudService()}
	provider, err := factory.new(CreateOptions{
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

func createTestVCSProvider(t *testing.T, db *pgdb, org *organization.Organization) *VCSProvider {
	provider := newTestVCSProvider(t, org)
	ctx := context.Background()

	err := db.create(ctx, provider)
	require.NoError(t, err)

	t.Cleanup(func() {
		db.delete(ctx, provider.ID)
	})
	return provider
}

type fakeService struct {
	provider *VCSProvider

	Service
}

func (f *fakeService) CreateVCSProvider(ctx context.Context, opts CreateOptions) (*VCSProvider, error) {
	return f.provider, nil
}

func (f *fakeService) list(context.Context, string) ([]*VCSProvider, error) {
	return []*VCSProvider{f.provider}, nil
}

func (f *fakeService) delete(context.Context, string) (*VCSProvider, error) {
	return f.provider, nil
}
