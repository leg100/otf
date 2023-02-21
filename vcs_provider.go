package otf

import (
	"context"
	"time"

	"github.com/leg100/otf/cloud"
)

type VCSProvider struct {
	ID           string
	String       string
	Token        string
	CreatedAt    time.Time
	Name         string
	Organization string
	CloudConfig  cloud.Config
}

// VCSProviderService provides access to vcs providers
type VCSProviderApp interface {
	GetVCSProvider(ctx context.Context, id string) (VCSProvider, error)
	ListVCSProviders(ctx context.Context, organization string) ([]VCSProvider, error)

	// GetVCSClient combines retrieving a vcs provider and construct a cloud
	// client from that provider.
	//
	// TODO: rename vcs provider to cloud client; the central purpose of the vcs
	// provider is, after all, to construct a cloud client.
	GetVCSClient(ctx context.Context, providerID string) (cloud.Client, error)
}
