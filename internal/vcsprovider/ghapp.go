package vcsprovider

import "github.com/leg100/otf/internal"

type (
	GithubApp struct {
		ID             string
		AppID          int64
		InstallationID int64
		WebhookSecret  string
		PrivateKey     string
	}

	CreateGithubAppOptions struct {
		AppID          int64
		InstallationID int64
		WebhookSecret  string
		PrivateKey     string
	}
)

func newGithubApp(opts CreateGithubAppOptions) *GithubApp {
	return &GithubApp{
		ID:             internal.NewID("github-app"),
		AppID:          opts.AppID,
		InstallationID: opts.InstallationID,
		WebhookSecret:  opts.WebhookSecret,
		PrivateKey:     opts.PrivateKey,
	}
}
