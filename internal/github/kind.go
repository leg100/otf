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
	apiURL              *internal.WebURL
	service             *Service
	skipTLSVerification bool
}

func registerVCSKinds(
	svc *Service,
	vcsService *vcs.Service,
	apiURL *internal.WebURL,
	skipTLSVerification bool,
) {
	provider := &kindProvider{
		service:             svc,
		db:                  svc.db,
		apiURL:              apiURL,
		skipTLSVerification: skipTLSVerification,
	}
	vcsService.RegisterKind(vcs.Kind{
		ID:            AppKindID,
		Icon:          Icon(),
		DefaultAPIURL: apiURL,
		AppKind:       provider,
		NewClient:     provider.NewClient,
		// Github apps don't need webhooks on repositories.
		SkipRepohook: true,
		Source:       internal.Ptr(Source),
	})
	vcsService.RegisterKind(vcs.Kind{
		ID:            TokenKindID,
		Icon:          Icon(),
		DefaultAPIURL: apiURL,
		TokenKind: &vcs.TokenKind{
			Description: tokenDescription(apiURL.Host),
		},
		NewClient:    provider.NewClient,
		EventHandler: HandleEvent,
		// Github token kind vcs providers can be created via the TFE API.
		TFEServiceProvider: vcs.ServiceProviderGithub,
		Source:             internal.Ptr(Source),
	})
}

func (p *kindProvider) NewClient(ctx context.Context, cfg vcs.ClientConfig) (vcs.Client, error) {
	opts := ClientOptions{
		APIURL:              p.apiURL,
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
