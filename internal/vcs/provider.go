package vcs

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/resource"
)

type (
	// Provider is an identifiable client for interacting with a VCS host.
	Provider struct {
		ID           resource.TfeID
		Name         string
		CreatedAt    time.Time
		Organization organization.Name
		Token        *string
		Installation *Installation
		// The kind of provider
		Kind
		// Client is the actual client for interacting with the VCS host.
		Client
		// The base URL of the the provider.
		BaseURL *internal.WebURL
		// NOTE: OTF doesn't use these fields but they're persisted in order to
		// satisfy the go-tfe integration tests and/or the tfe API.
		apiURL              *internal.WebURL
		serviceProviderType TFEServiceProviderType
	}

	// factory produces providers
	factory struct {
		kinds *kindDB
	}

	CreateOptions struct {
		Organization organization.Name `schema:"organization_name,required"`
		Name         string
		KindID       KindID `schema:"kind,required"`
		// Token and InstallID are mutually exclusive.
		Token     *string
		InstallID *int64 `schema:"install_id"`
		// Optional.
		BaseURL *internal.WebURL `schema:"base_url,required"`
		// Optional.
		tfeServiceProviderType *TFEServiceProviderType
		// Optional.
		apiURL *internal.WebURL
	}

	UpdateOptions struct {
		Name    string
		Token   *string
		BaseURL *internal.WebURL
	}

	ListOptions struct {
		resource.PageOptions
		Organization organization.Name `schema:"organization_name"`
	}
)

func (f *factory) newProvider(ctx context.Context, opts CreateOptions) (*Provider, error) {
	kind, err := f.kinds.GetKind(opts.KindID)
	if err != nil {
		return nil, err
	}
	provider := &Provider{
		ID:           resource.NewTfeID(resource.VCSProviderKind),
		Name:         opts.Name,
		CreatedAt:    internal.CurrentTimestamp(nil),
		Organization: opts.Organization,
		Kind:         kind,
		BaseURL:      opts.BaseURL,
	}
	if kind.AppKind != nil {
		if opts.InstallID == nil {
			return nil, errors.New("install ID required for client")
		}
		app, err := kind.AppKind.GetApp(ctx)
		if err != nil {
			return nil, err
		}
		install, err := app.GetInstallation(ctx, *opts.InstallID)
		if err != nil {
			return nil, err
		}
		provider.Token = opts.Token
		provider.Installation = &install
	} else if kind.TokenKind != nil {
		if opts.Token == nil {
			return nil, errors.New("token required for client")
		}
		provider.Token = opts.Token
	} else {
		return nil, errors.New("an installation or a token must be specified")
	}
	// If caller hasn't specified a base URL then use the kind's default.
	if opts.BaseURL != nil {
		provider.BaseURL = opts.BaseURL
	} else {
		provider.BaseURL = kind.DefaultURL
	}
	// If caller hasn't specified a TFE service provider type to assign to the
	// provider then retrieve the first type that the kind supports.
	if opts.tfeServiceProviderType == nil {
		provider.serviceProviderType = kind.TFEServiceProviders[0]
	} else {
		provider.serviceProviderType = *opts.tfeServiceProviderType
	}
	// If caller hasn't specified an apiURL then set it to the same value as the
	// BaseURL
	if opts.apiURL == nil {
		provider.apiURL = provider.BaseURL
	} else {
		provider.apiURL = opts.apiURL
	}
	client, err := kind.NewClient(ctx, ClientConfig{
		Token:        provider.Token,
		Installation: provider.Installation,
		BaseURL:      provider.BaseURL,
	})
	if err != nil {
		return nil, err
	}
	provider.Client = client
	return provider, nil
}

// String provides a human meaningful description of the vcs provider.
func (t *Provider) String() string {
	if t.Name != "" {
		return t.Name
	}
	return internal.Title(string(t.Kind.ID))
}

func (t *Provider) Update(opts UpdateOptions) error {
	if opts.Token != nil {
		// If token is set it cannot be empty
		if *opts.Token == "" {
			return fmt.Errorf("token: %w", internal.ErrEmptyValue)
		}
		t.Token = opts.Token
	}
	if opts.BaseURL != nil {
		t.BaseURL = opts.BaseURL
	}
	t.Name = opts.Name
	return nil
}

// LogValue implements slog.LogValuer.
func (t *Provider) LogValue() slog.Value {
	attrs := []slog.Attr{
		slog.String("id", t.ID.String()),
		slog.Any("organization", t.Organization),
		slog.String("name", t.String()),
		slog.String("kind", string(t.Kind.ID)),
		slog.Any("api_url", t.BaseURL),
	}
	if t.Installation != nil {
		attrs = append(attrs, slog.Int64("install_id", t.Installation.ID))
		attrs = append(attrs, slog.Int64("install_app_id", t.Installation.AppID))
	}
	return slog.GroupValue(attrs...)
}
