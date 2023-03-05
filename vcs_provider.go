package otf

import (
	"context"
	"time"

	"github.com/leg100/otf/cloud"
)

type (
	// VCSProvider provides authenticated access to a VCS. Equivalent to an OAuthClient in
	// TFE.
	VCSProvider struct {
		ID           string
		CreatedAt    time.Time
		Name         string       // TODO: rename to description (?)
		CloudConfig  cloud.Config // cloud config for creating client
		Token        string       // credential for creating client
		Organization string       // vcs provider belongs to an organization
	}

	// VCSProviderService provides access to vcs providers
	VCSProviderService interface {
		GetVCSProvider(ctx context.Context, id string) (*VCSProvider, error)
		ListVCSProviders(ctx context.Context, organization string) ([]*VCSProvider, error)

		// GetVCSClient combines retrieving a vcs provider and construct a cloud
		// client from that provider.
		//
		// TODO: rename vcs provider to cloud client; the central purpose of the vcs
		// provider is, after all, to construct a cloud client.
		GetVCSClient(ctx context.Context, providerID string) (cloud.Client, error)
	}

	// VCSProviderStore persists vcs providers
	VCSProviderStore interface {
		CreateVCSProvider(ctx context.Context, provider *VCSProvider) error
		GetVCSProvider(ctx context.Context, id string) (*VCSProvider, error)
		ListVCSProviders(ctx context.Context, organization string) ([]*VCSProvider, error)
		DeleteVCSProvider(ctx context.Context, id string) error
	}
)

func (t *VCSProvider) String() string { return t.Name }

func (t *VCSProvider) NewClient(ctx context.Context) (cloud.Client, error) {
	return t.CloudConfig.NewClient(ctx, cloud.Credentials{
		PersonalToken: &t.Token,
	})
}

func (t *VCSProvider) MarshalLog() any {
	return struct {
		ID           string `json:"id"`
		Organization string `json:"organization"`
		Name         string `json:"name"`
		Cloud        string `json:"cloud"`
	}{
		ID:           t.ID,
		Organization: t.Organization,
		Name:         t.Name,
		Cloud:        t.CloudConfig.Name,
	}
}
