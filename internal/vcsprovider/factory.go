package vcsprovider

import (
	"fmt"
	"time"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/github"
)

type (
	// factory makes vcs providers
	factory struct {
		CloudService
	}

	CreateOptions struct {
		Organization string
		Name         string
		Cloud        string
		ID           *string
		CreatedAt    *time.Time

		Token     *string
		GithubApp *github.Install // mutually exclusive with Token.
	}
)

func (f *factory) new(opts CreateOptions) (*VCSProvider, error) {
	cloudConfig, err := f.GetCloudConfig(opts.Cloud)
	if err != nil {
		return nil, err
	}

	if opts.Token == nil && opts.GithubApp == nil {
		return nil, &internal.MissingParameterError{Parameter: "must specify either token or github-app"}
	}
	if opts.Token != nil && opts.GithubApp != nil {
		return nil, fmt.Errorf("cannot specify both token and github app")
	}

	provider := &VCSProvider{
		ID:           internal.NewID("vcs"),
		CreatedAt:    internal.CurrentTimestamp(),
		Name:         opts.Name,
		Organization: opts.Organization,
		CloudConfig:  cloudConfig,
		Token:        opts.Token,
		GithubApp:    opts.GithubApp,
	}
	if opts.ID != nil {
		provider.ID = *opts.ID
	}
	if opts.CreatedAt != nil {
		provider.CreatedAt = *opts.CreatedAt
	}
	return provider, nil
}
