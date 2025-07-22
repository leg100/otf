package github

import (
	"context"
	"fmt"

	"github.com/a-h/templ"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/vcs"
	"golang.org/x/exp/slog"
)

type (
	// App is a Github App. Includes a client for interacting with Github
	// authenticated as the app.
	App struct {
		ID            AppID  `db:"github_app_id"` // github's app id
		Slug          string // github's "slug" name
		WebhookSecret string `db:"webhook_secret"`
		PrivateKey    string `db:"private_key"`

		// Organization is the name of the *github* organization that owns the
		// app. If the app is owned by a user then this is nil.
		Organization *string `db:"organization"`
		GithubURL    *internal.WebURL

		*Client
	}

	CreateAppOptions struct {
		BaseURL             *internal.WebURL
		AppID               int64
		WebhookSecret       string
		PrivateKey          string
		Slug                string
		Organization        *string
		SkipTLSVerification bool
	}
)

func newApp(opts CreateAppOptions) (*App, error) {
	app := &App{
		ID:            AppID(opts.AppID),
		Slug:          opts.Slug,
		WebhookSecret: opts.WebhookSecret,
		PrivateKey:    opts.PrivateKey,
		Organization:  opts.Organization,
	}

	client, err := NewClient(ClientOptions{
		BaseURL:             opts.BaseURL,
		SkipTLSVerification: opts.SkipTLSVerification,
		AppCredentials: &AppCredentials{
			ID:         app.ID,
			PrivateKey: app.PrivateKey,
		},
	})
	if err != nil {
		return nil, err
	}
	app.Client = client

	return app, nil
}

func (a *App) String() string { return a.Slug }

// URL returns the app's URL on GitHub
func (a *App) URL(hostname string) string {
	return "https://" + hostname + "/apps/" + a.Slug
}

// NewInstallURL returns the GitHub URL for creating a new install of the app.
func (a *App) NewInstallURL(hostname string) string {
	return "https://" + hostname + "/apps/" + a.Slug + "/installations/new"
}

func (a *App) InstallationLink() templ.SafeURL {
	return templ.SafeURL(a.NewInstallURL(a.GithubURL.Host))
}

// LogValue implements slog.LogValuer.
func (a *App) LogValue() slog.Value {
	return slog.GroupValue(
		slog.Any("id", a.ID),
		slog.String("slug", a.Slug),
	)
}

// AdvancedURL returns the URL for the "advanced" settings on github
func (a *App) AdvancedURL() templ.SafeURL {
	path := fmt.Sprintf("/settings/apps/%s/advanced", a.Slug)
	if a.Organization != nil {
		path = fmt.Sprintf("/organizations/%s%s", *a.Organization, path)
	}
	u := *a.GithubURL
	u.Path = path
	return templ.SafeURL(u.String())
}

// ListInstallations lists installations of the currently authenticated app.
func (a *App) ListInstallations(ctx context.Context) ([]vcs.Installation, error) {
	installs, resp, err := a.client.Apps.ListInstallations(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	to := make([]vcs.Installation, len(installs))
	for i, install := range installs {
		var err error
		to[i], err = vcs.NewInstallation(install)
		if err != nil {
			return nil, err
		}
	}

	return to, err
}

func (a *App) GetInstallation(ctx context.Context, installID int64) (vcs.Installation, error) {
	install, resp, err := a.client.Apps.GetInstallation(ctx, installID)
	if err != nil {
		return vcs.Installation{}, err
	}
	defer resp.Body.Close()

	return vcs.NewInstallation(install)
}

// DeleteInstallation deletes an installation of a github app with the given
// installation ID.
func (a *App) DeleteInstallation(ctx context.Context, installID int64) error {
	resp, err := a.client.Apps.DeleteInstallation(ctx, installID)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return err
}
