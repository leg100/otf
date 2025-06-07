package github

import (
	"context"

	"github.com/a-h/templ"
	"github.com/leg100/otf/internal/vcs"
)

const (
	TokenKind vcs.KindID = "github-token"
	AppKind   vcs.KindID = "github-app"
)

type provider struct {
	db                  *pgdb
	hostname            string
	service             *Service
	skipTLSVerification bool
}

func registerProviders(
	svc *Service,
	vcsService *vcs.Service,
	hostname string,
	skipTLSVerification bool,
) {
	provider := &provider{
		service:             svc,
		db:                  svc.db,
		hostname:            hostname,
		skipTLSVerification: skipTLSVerification,
	}
	vcsService.RegisterKind(vcs.Kind{
		ID:               AppKind,
		Name:             "GitHub (App)",
		Icon:             Icon(),
		Hostname:         hostname,
		InstallationKind: provider,
		NewClient:        provider.NewClient,
		// Github apps don't need webhooks on repositories.
		SkipRepohook: true,
	})
	vcsService.RegisterKind(vcs.Kind{
		ID:       TokenKind,
		Name:     "GitHub (Token)",
		Icon:     Icon(),
		Hostname: hostname,
		TokenKind: &vcs.TokenKind{
			Description: tokenDescription(hostname),
		},
		NewClient: provider.NewClient,
		// Github token kind vcs providers can be created via the TFE API.
		TFEServiceProvider: vcs.ServiceProviderGithub,
	})
}

func (p *provider) NewClient(ctx context.Context, cfg vcs.Config) (vcs.Client, error) {
	opts := ClientOptions{
		Hostname:            p.hostname,
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

func (p *provider) GetInstallation(ctx context.Context, id int64) (vcs.Installation, error) {
	creds, err := p.service.GetInstallCredentials(ctx, id)
	if err != nil {
		return vcs.Installation{}, err
	}
	install := vcs.Installation{
		ID:           creds.ID,
		AppID:        int64(creds.AppCredentials.ID),
		Organization: creds.Organization,
		Username:     creds.User,
	}
	return install, nil
}

func (p *provider) ListInstallations(ctx context.Context) (vcs.ListInstallationsResult, error) {
	app, err := p.service.GetApp(ctx)
	if err != nil {
		return vcs.ListInstallationsResult{}, err
	}
	installs, err := p.service.ListInstallations(ctx)
	if err != nil {
		return vcs.ListInstallationsResult{}, err
	}
	m := make(map[string]int64, len(installs))
	for _, install := range installs {
		m[install.String()] = *install.ID
	}
	result := vcs.ListInstallationsResult{
		InstallationLink: templ.SafeURL(app.NewInstallURL(p.hostname)),
	}
	return result, nil
}
