package vcsprovider

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/cloud"
	"github.com/leg100/otf/internal/organization"
	"github.com/stretchr/testify/require"
)

func newTestVCSProvider(t *testing.T, org *organization.Organization) *VCSProvider {
	provider, err := newProvider(context.Background(), newOptions{
		CreateOptions: CreateOptions{
			Organization: org.Name,
			// unit tests require a legitimate cloud name to avoid invalid foreign
			// key error upon insert/update
			Kind:  cloud.GithubKind,
			Name:  uuid.NewString(),
			Token: internal.String("token"),
		},
		cloudHostnames: map[cloud.Kind]string{
			cloud.GithubKind: "example.com",
		},
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
