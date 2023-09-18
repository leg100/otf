// Package vcsprovider is responsible for VCS providers
package vcsprovider

import (
	"context"
	"errors"
	"fmt"
	"time"

	"log/slog"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/cloud"
	"github.com/leg100/otf/internal/github"
)

type (
	// VCSProvider provides authenticated access to a VCS. Equivalent to an OAuthClient in
	// TFE.
	VCSProvider struct {
		ID           string
		CreatedAt    time.Time
		CloudConfig  cloud.Config // cloud config for creating client
		Organization string       // vcs provider belongs to an organization
		Name         string

		Token     *string         // personal access token.
		GithubApp *github.Install // mutually exclusive with Token.
	}

	CreateOptions struct {
		Organization string
		Cloud        string
		ID           *string
		Name         string
		CreatedAt    *time.Time

		Token              *string
		GithubAppInstallID *int64

		CloudService
		github.GithubAppService
	}

	UpdateOptions struct {
		Token *string
		Name  string
	}
)

func newProvider(ctx context.Context, opts CreateOptions) (*VCSProvider, error) {
	cloudConfig, err := opts.GetCloudConfig(opts.Cloud)
	if err != nil {
		return nil, err
	}
	provider := &VCSProvider{
		ID:           internal.NewID("vcs"),
		CreatedAt:    internal.CurrentTimestamp(),
		Organization: opts.Organization,
		CloudConfig:  cloudConfig,
		Name:         opts.Name,
	}
	if opts.ID != nil {
		provider.ID = *opts.ID
	}
	if opts.CreatedAt != nil {
		provider.CreatedAt = *opts.CreatedAt
	}
	if opts.Token != nil {
		if err := provider.setToken(*opts.Token); err != nil {
			return nil, err
		}
	} else if opts.GithubAppInstallID != nil {
		app, err := opts.GetGithubApp(ctx)
		if err != nil {
			return nil, err
		}
		provider.GithubApp = &github.Install{
			ID:  *opts.GithubAppInstallID,
			App: app,
		}
	} else {
		return nil, errors.New("must specify either token or github app installation ID")
	}
	return provider, nil
}

// String provides a human meaningful description of the vcs provider, using the
// name if set, otherwise using the name of the underlying cloud provider.
func (t *VCSProvider) String() string {
	if t.Name != "" {
		return t.Name
	}
	return t.CloudConfig.Name
}

func (t *VCSProvider) NewClient(ctx context.Context) (cloud.Client, error) {
	return t.CloudConfig.NewClient(ctx, cloud.Credentials{
		PersonalToken: t.Token,
	})
}

func (t *VCSProvider) Update(opts UpdateOptions) error {
	t.Name = opts.Name
	if opts.Token != nil {
		if err := t.setToken(*opts.Token); err != nil {
			return err
		}
	}
	return nil
}

// LogValue implements slog.LogValuer.
func (t *VCSProvider) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("id", t.ID),
		slog.String("organization", t.Organization),
		slog.String("name", t.String()),
		slog.String("cloud", t.CloudConfig.Name),
	)
}

func (t *VCSProvider) setToken(token string) error {
	if token == "" {
		return fmt.Errorf("token: %w", internal.ErrEmptyValue)
	}
	t.Token = &token
	return nil
}
