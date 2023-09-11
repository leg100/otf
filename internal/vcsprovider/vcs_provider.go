// Package vcsprovider is responsible for VCS providers
package vcsprovider

import (
	"context"
	"fmt"
	"time"

	"log/slog"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/cloud"
)

type (
	// VCSProvider provides authenticated access to a VCS. Equivalent to an OAuthClient in
	// TFE.
	VCSProvider struct {
		ID           string
		CreatedAt    time.Time
		CloudConfig  cloud.Config // cloud config for creating client
		Token        string       // credential for creating client
		Organization string       // vcs provider belongs to an organization
		Name         string
	}

	CreateOptions struct {
		Organization string
		Token        string
		Cloud        string
		ID           *string
		Name         string
		CreatedAt    *time.Time
	}

	UpdateOptions struct {
		Token *string
		Name  string
	}
)

func newProvider(cloudService CloudService, opts CreateOptions) (*VCSProvider, error) {
	cloudConfig, err := cloudService.GetCloudConfig(opts.Cloud)
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
		PersonalToken: &t.Token,
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
	t.Token = token
	return nil
}
