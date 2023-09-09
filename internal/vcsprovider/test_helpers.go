package vcsprovider

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/inmem"
	"github.com/leg100/otf/internal/organization"
	"github.com/stretchr/testify/require"
)

func newTestVCSProvider(t *testing.T, org *organization.Organization) *VCSProvider {
	cloudService := inmem.NewCloudServiceWithDefaults()
	provider, err := newProvider(cloudService, CreateOptions{
		Organization: org.Name,
		// unit tests require a legitimate cloud name to avoid invalid foreign
		// key error upon insert/update
		Cloud: "github",
		Name:  internal.String(uuid.NewString()),
		Token: uuid.NewString(),
	})
	require.NoError(t, err)
	return provider
}

type fakeService struct {
	provider *VCSProvider

	Service
}

func (f *fakeService) CreateVCSProvider(ctx context.Context, opts CreateOptions) (*VCSProvider, error) {
	return f.provider, nil
}

func (f *fakeService) ListVCSProviders(context.Context, string) ([]*VCSProvider, error) {
	return []*VCSProvider{f.provider}, nil
}

func (f *fakeService) DeleteVCSProvider(context.Context, string) (*VCSProvider, error) {
	return f.provider, nil
}
