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

		// The base URL of the API of the provider.
		APIURL *internal.WebURL
		// NOTE: OTF doesn't use this field but it's persisted in order to
		// satisfy the go-tfe integration tests.
		httpURL *internal.WebURL
	}

	// factory produces providers
	factory struct {
		kinds *kindDB
	}

	CreateOptions struct {
		Organization organization.Name `schema:"organization_name,required"`
		Name         string
		KindID       KindID `schema:"kind,required"`
		Token        *string
		InstallID    *int64           `schema:"install_id"`
		APIURL       *internal.WebURL `schema:"api_url,required"`
		HTTPURL      *internal.WebURL
	}

	UpdateOptions struct {
		Name   string
		Token  *string
		APIURL *internal.WebURL
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
		httpURL:      opts.HTTPURL,
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
	if opts.APIURL != nil {
		provider.APIURL = opts.APIURL
	} else {
		provider.APIURL = kind.DefaultAPIURL
	}
	client, err := kind.NewClient(ctx, ClientConfig{
		Token:        provider.Token,
		Installation: provider.Installation,
		APIURL:       provider.APIURL,
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
	if opts.APIURL != nil {
		t.APIURL = opts.APIURL
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
		slog.Any("api_url", t.APIURL),
	}
	if t.Installation != nil {
		attrs = append(attrs, slog.Int64("install_id", t.Installation.ID))
		attrs = append(attrs, slog.Int64("install_app_id", t.Installation.AppID))
	}
	return slog.GroupValue(attrs...)
}
