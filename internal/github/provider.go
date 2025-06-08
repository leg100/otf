package github

import (
	"context"

	"github.com/a-h/templ"
	"github.com/leg100/otf/internal/vcs"
)

const (
	TokenKindID vcs.KindID = "github-token"
	AppKindID   vcs.KindID = "github-app"
)

type provider struct {
	db                  *pgdb
	hostname            string
	service             *Service
	skipTLSVerification bool
}

func registerVCSKinds(
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
		ID:               AppKindID,
		Name:             "GitHub (App)",
		Source:           "github",
		Icon:             Icon(),
		Hostname:         hostname,
		InstallationKind: provider,
		NewClient:        provider.NewClient,
		// Github apps don't need webhooks on repositories.
		SkipRepohook: true,
	})
	vcsService.RegisterKind(vcs.Kind{
		ID:       TokenKindID,
		Name:     "GitHub (Token)",
		Source:   "github",
		Icon:     Icon(),
		Hostname: hostname,
		TokenKind: &vcs.TokenKind{
			Description: tokenDescription(hostname),
		},
		NewClient:    provider.NewClient,
		EventHandler: HandleEvent,
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
	install, err := p.service.GetInstallation(ctx, id)
	if err != nil {
		return vcs.Installation{}, err
	}
	vcsInstall := vcs.Installation{
		ID:           *install.ID,
		AppID:        int64(*install.AppID),
		Organization: install.Organization(),
		Username:     install.Username(),
	}
	return vcsInstall, nil
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
	vcsInstalls := make([]vcs.Installation, len(installs))
	for i, install := range installs {
		vcsInstalls[i] = vcs.Installation{
			ID:           *install.ID,
			AppID:        *install.AppID,
			Username:     install.Username(),
			Organization: install.Organization(),
		}
	}
	result := vcs.ListInstallationsResult{
		InstallationLink: templ.SafeURL(app.NewInstallURL(p.hostname)),
		Results:          vcsInstalls,
	}
	return result, nil
}
