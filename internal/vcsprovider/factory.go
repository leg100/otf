package vcsprovider

import (
	"time"

	"github.com/leg100/otf/internal"
)

// factory makes vcs providers
type factory struct {
	CloudService
}

func (f *factory) new(opts CreateOptions) (*VCSProvider, error) {
	cloudConfig, err := f.GetCloudConfig(opts.Cloud)
	if err != nil {
		return nil, err
	}

	provider := &VCSProvider{
		ID:           internal.NewID("vcs"),
		CreatedAt:    internal.CurrentTimestamp(),
		Name:         opts.Name,
		Organization: opts.Organization,
		CloudConfig:  cloudConfig,
		Token:        opts.Token,
	}
	if opts.ID != nil {
		provider.ID = *opts.ID
	}
	if opts.CreatedAt != nil {
		provider.CreatedAt = *opts.CreatedAt
	}
	return provider, nil
}

type CreateOptions struct {
	Organization string
	Token        string
	Name         string
	Cloud        string
	ID           *string
	CreatedAt    *time.Time
}
