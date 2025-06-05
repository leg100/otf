// Package vcsprovider is responsible for VCS providers
package vcsprovider

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/vcs"
)

type (
	// VCSProvider provides authenticated access to a VCS.
	VCSProvider struct {
		ID           resource.TfeID
		Name         string
		CreatedAt    time.Time
		Organization organization.Name // name of OTF organization

		// TODO: this should not be part of the vcs provider
		Hostname string // hostname of github/gitlab etc

		Kind vcs.Kind // github/gitlab etc.

		Config
		vcs.Client

		// TODO: this should not be part of the vcs provider
		skipTLSVerification bool // toggle skipping verification of VCS host's TLS cert.
	}

	// factory produces VCS providers
	factory struct {
		schemas map[vcs.Kind]ConfigSchema

		skipTLSVerification bool // toggle skipping verification of VCS host's TLS cert.
	}

	CreateOptions struct {
		Organization organization.Name `schema:"organization_name,required"`
		Name         string
		Kind         vcs.Kind
		Token        *string
		InstallID    *int64
	}

	UpdateOptions struct {
		Name  string
		Token *string
	}

	ListOptions struct {
		resource.PageOptions
		Organization organization.Name `schema:"organization_name"`
	}
)

func (f *factory) newProvider(ctx context.Context, opts CreateOptions) (*VCSProvider, error) {
	provider := &VCSProvider{
		ID:                  resource.NewTfeID(resource.VCSProviderKind),
		Name:                opts.Name,
		CreatedAt:           internal.CurrentTimestamp(nil),
		Organization:        opts.Organization,
		Kind:                opts.Kind,
		skipTLSVerification: f.skipTLSVerification,
	}
	schema, ok := f.schemas[opts.Kind]
	if !ok {
		return nil, errors.New("schema not found")
	}
	provider.Hostname = schema.Hostname
	var cfg Config
	if schema.WantsInstallation {
		if opts.InstallID == nil {
			return nil, errors.New("install ID required for client")
		}
		install, err := schema.GetInstallation(ctx, *opts.InstallID)
		if err != nil {
			return nil, err
		}
		cfg = Config{Installation: &install}
	} else if schema.WantsToken {
		if opts.Token == nil {
			return nil, errors.New("token required for client")
		}
		cfg = Config{Token: opts.Token}
	}
	client, err := schema.NewClient(cfg)
	if err != nil {
		return nil, err
	}
	provider.Client = client
	return provider, nil
}

// String provides a human meaningful description of the vcs provider, using the
// name if set; otherwise a name is constructed using both the underlying cloud
// kind and the auth kind.
func (t *VCSProvider) String() string {
	return string(t.Kind)
}

func (t *VCSProvider) Update(opts UpdateOptions) error {
	if opts.Token != nil {
		// If token is set it cannot be empty
		if *opts.Token == "" {
			return fmt.Errorf("token: %w", internal.ErrEmptyValue)
		}
		opts.Token = opts.Token
	}
	t.Name = opts.Name
	return nil
}

// LogValue implements slog.LogValuer.
func (t *VCSProvider) LogValue() slog.Value {
	attrs := []slog.Attr{
		slog.String("id", t.ID.String()),
		slog.Any("organization", t.Organization),
		slog.String("name", t.String()),
		slog.String("kind", string(t.Kind)),
	}
	return slog.GroupValue(attrs...)
}
