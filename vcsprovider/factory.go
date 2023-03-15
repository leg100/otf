package vcsprovider

import (
	"time"

	"github.com/leg100/otf"
	"github.com/leg100/otf/cloud"
)

// factory makes vcs providers
type factory struct {
	cloud.Service
}

func (f *factory) new(opts createOptions) (*VCSProvider, error) {
	cloudConfig, err := f.GetCloudConfig(opts.Cloud)
	if err != nil {
		return nil, err
	}

	provider := &VCSProvider{
		ID:           otf.NewID("vcs"),
		CreatedAt:    otf.CurrentTimestamp(),
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

type createOptions struct {
	Organization string
	Token        string
	Name         string
	Cloud        string
	ID           *string
	CreatedAt    *time.Time
}
