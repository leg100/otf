package vcsprovider

import (
	"context"
	"time"

	"github.com/leg100/otf/cloud"
)

// VCSProvider provides authenticated access to a VCS. Equivalent to an OAuthClient in
// TFE.
type VCSProvider struct {
	id           string
	createdAt    time.Time
	name         string       // TODO: rename to description (?)
	cloudConfig  cloud.Config // cloud config for creating client
	token        string       // credential for creating client
	organization string       // vcs provider belongs to an organization
}

func (t *VCSProvider) ID() string                { return t.id }
func (t *VCSProvider) String() string            { return t.name }
func (t *VCSProvider) Token() string             { return t.token }
func (t *VCSProvider) CreatedAt() time.Time      { return t.createdAt }
func (t *VCSProvider) Name() string              { return t.name }
func (t *VCSProvider) Organization() string      { return t.organization }
func (t *VCSProvider) CloudConfig() cloud.Config { return t.cloudConfig }

func (t *VCSProvider) NewClient(ctx context.Context) (cloud.Client, error) {
	return t.cloudConfig.NewClient(ctx, cloud.Credentials{
		PersonalToken: &t.token,
	})
}

func (t *VCSProvider) MarshalLog() any {
	return struct {
		ID           string `json:"id"`
		Organization string `json:"organization"`
		Name         string `json:"name"`
		Cloud        string `json:"cloud"`
	}{
		ID:           t.id,
		Organization: t.organization,
		Name:         t.name,
		Cloud:        t.cloudConfig.Name,
	}
}

// VCSProviderService provides access to vcs providers
type VCSProviderService interface {
	CreateVCSProvider(ctx context.Context, opts createOptions) (*VCSProvider, error)
	GetVCSProvider(ctx context.Context, id string) (*VCSProvider, error)
	ListVCSProviders(ctx context.Context, organization string) ([]*VCSProvider, error)
	DeleteVCSProvider(ctx context.Context, id string) (*VCSProvider, error)

	// GetVCSClient combines retrieving a vcs provider and construct a cloud
	// client from that provider.
	//
	// TODO: rename vcs provider to cloud client; the central purpose of the vcs
	// provider is, after all, to construct a cloud client.
	GetVCSClient(ctx context.Context, providerID string) (cloud.Client, error)
}

// VCSProviderStore persists vcs providers
type VCSProviderStore interface {
	CreateVCSProvider(ctx context.Context, provider *VCSProvider) error
	GetVCSProvider(ctx context.Context, id string) (*VCSProvider, error)
	ListVCSProviders(ctx context.Context, organization string) ([]*VCSProvider, error)
	DeleteVCSProvider(ctx context.Context, id string) error
}
