package github

import (
	"fmt"

	"golang.org/x/exp/slog"
)

type (
	App struct {
		ID            AppID  `db:"github_app_id"` // github's app id
		Slug          string // github's "slug" name
		WebhookSecret string `db:"webhook_secret"`
		PrivateKey    string `db:"private_key"`

		// Organization is the name of the *github* organization that owns the
		// app. If the app is owned by a user then this is nil.
		Organization *string `db:"organization"`
	}

	CreateAppOptions struct {
		AppID         int64
		WebhookSecret string
		PrivateKey    string
		Slug          string
		Organization  *string
	}
)

func newApp(opts CreateAppOptions) *App {
	return &App{
		ID:            AppID(opts.AppID),
		Slug:          opts.Slug,
		WebhookSecret: opts.WebhookSecret,
		PrivateKey:    opts.PrivateKey,
		Organization:  opts.Organization,
	}
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

// LogValue implements slog.LogValuer.
func (a *App) LogValue() slog.Value {
	return slog.GroupValue(
		slog.Any("id", a.ID),
		slog.String("slug", a.Slug),
	)
}

// AdvancedURL returns the URL for the "advanced" settings on github
func (a *App) AdvancedURL(hostname string) string {
	path := fmt.Sprintf("/settings/apps/%s/advanced", a.Slug)
	if a.Organization != nil {
		path = fmt.Sprintf("/organizations/%s%s", *a.Organization, path)
	}
	return fmt.Sprintf("https://%s%s", hostname, path)
}
