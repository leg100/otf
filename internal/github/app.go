package github

import "github.com/leg100/otf/internal"

type (
	GithubApp struct {
		ID            string
		AppID         int64
		WebhookSecret string
		PrivateKey    string
	}

	Install struct {
		InstallID int64
		GithubApp
	}

	newGithubAppOptions struct {
		AppID         int64
		WebhookSecret string
		PrivateKey    string
	}
)

func newGithubApp(opts newGithubAppOptions) *GithubApp {
	return &GithubApp{
		ID:            internal.NewID("ghapp"),
		AppID:         opts.AppID,
		WebhookSecret: opts.WebhookSecret,
		PrivateKey:    opts.PrivateKey,
	}
}
