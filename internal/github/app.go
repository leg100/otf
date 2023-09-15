package github

import "github.com/leg100/otf/internal"

type (
	GithubApp struct {
		ID            int64 // github's app id
		WebhookSecret string
		PrivateKey    string
	}

	Install struct {
		ID       string // OTF's install ID
		GithubID int64  // github's install ID
		App      *GithubApp
	}
)

func newApp(opts CreateAppOptions) *GithubApp {
	return &GithubApp{
		ID:            opts.AppID,
		WebhookSecret: opts.WebhookSecret,
		PrivateKey:    opts.PrivateKey,
	}
}
func newInstall(installID int64, app *GithubApp) Install {
	return Install{
		ID:       internal.NewID("ghain"),
		GithubID: installID,
		App:      app,
	}
}
