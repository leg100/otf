package vcsprovider

import (
	"time"

	"github.com/leg100/otf/internal"
)

type (
	// factory makes vcs providers
	factory struct {
		CloudService
	}

	CreateOptions struct {
		Organization string
		Token        string
		Name         string
		Cloud        string
		ID           *string
		CreatedAt    *time.Time
	}

	UpdateOptions struct {
		Token *string
		Name  *string
	}
)

func (f *factory) new(opts CreateOptions) (*VCSProvider, error) {
	cloudConfig, err := f.GetCloudConfig(opts.Cloud)
	if err != nil {
		return nil, err
	}

	provider := &VCSProvider{
		ID:           internal.NewID("vcs"),
		CreatedAt:    internal.CurrentTimestamp(),
		Organization: opts.Organization,
		CloudConfig:  cloudConfig,
	}
	if err := provider.setName(opts.Name); err != nil {
		return nil, err
	}
	if err := provider.setToken(opts.Token); err != nil {
		return nil, err
	}
	if opts.ID != nil {
		provider.ID = *opts.ID
	}
	if opts.CreatedAt != nil {
		provider.CreatedAt = *opts.CreatedAt
	}
	return provider, nil
}
