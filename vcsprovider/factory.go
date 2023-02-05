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
		id:           otf.NewID("vcs"),
		createdAt:    otf.CurrentTimestamp(),
		name:         opts.Name,
		organization: opts.Organization,
		cloudConfig:  cloudConfig,
		token:        opts.Token,
	}
	if opts.ID != nil {
		provider.id = *opts.ID
	}
	if opts.CreatedAt != nil {
		provider.createdAt = *opts.CreatedAt
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
