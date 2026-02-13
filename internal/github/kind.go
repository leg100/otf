package github

import (
	"context"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/configversion/source"
	"github.com/leg100/otf/internal/vcs"
)

const (
	TokenKindID vcs.KindID    = "github-token"
	AppKindID   vcs.KindID    = "github-app"
	Source      source.Source = "github"
)

type kindProvider struct {
	db                  *appDB
	baseURL             *internal.WebURL
	service             *Service
	skipTLSVerification bool
}

func registerVCSKinds(
	svc *Service,
	vcsService *vcs.Service,
	defaultURL *internal.WebURL,
	skipTLSVerification bool,
) {
	provider := &kindProvider{
		service:             svc,
		db:                  svc.db,
		baseURL:             defaultURL,
		skipTLSVerification: skipTLSVerification,
	}
	vcsService.RegisterKind(vcs.Kind{
		ID:         AppKindID,
		Icon:       Icon(),
		DefaultURL: defaultURL,
		AppKind:    provider,
		NewClient:  provider.NewClient,
		// Github apps don't need webhooks on repositories.
		SkipRepohook: true,
		Source:       new(Source),
		TFEServiceProviders: []vcs.TFEServiceProviderType{
			vcs.ServiceProviderGithubApp,
		},
	})
	vcsService.RegisterKind(vcs.Kind{
		ID:         TokenKindID,
		Icon:       Icon(),
		DefaultURL: defaultURL,
		TokenKind: &vcs.TokenKind{
			Description: tokenDescription(defaultURL.Host),
		},
		NewClient:    provider.NewClient,
		EventHandler: HandleEvent,
		TFEServiceProviders: []vcs.TFEServiceProviderType{
			vcs.ServiceProviderGithub,
			vcs.ServiceProviderGithubEE,
		},
		Source: new(Source),
	})
}

func (p *kindProvider) NewClient(ctx context.Context, cfg vcs.ClientConfig) (vcs.Client, error) {
	opts := ClientOptions{
		BaseURL:             cfg.BaseURL,
		SkipTLSVerification: p.skipTLSVerification,
	}
	if cfg.Token != nil {
		opts.PersonalToken = cfg.Token
	} else if cfg.Installation != nil {
		app, err := p.db.get(ctx)
		if err != nil {
			return nil, err
		}
		opts.InstallCredentials = &InstallCredentials{
			ID:           cfg.Installation.ID,
			User:         cfg.Installation.Username,
			Organization: cfg.Installation.Organization,
			AppCredentials: AppCredentials{
				ID:         app.ID,
				PrivateKey: app.PrivateKey,
			},
		}
	}
	return NewClient(opts)
}

func (p *kindProvider) GetApp(ctx context.Context) (vcs.App, error) {
	return p.service.GetApp(ctx)
}
