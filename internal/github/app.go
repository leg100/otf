package github

import "github.com/leg100/otf/internal"

type (
	GithubApp struct {
		ID            string // OTF's app ID
		AppID         int64  // github's app ID
		WebhookSecret string
		PrivateKey    string
	}

	Install struct {
		ID        string // OTF's install ID
		InstallID int64  // github's install ID
		*GithubApp
	}
)

func newApp(opts CreateAppOptions) *GithubApp {
	return &GithubApp{
		ID:            internal.NewID("ghapp"),
		AppID:         opts.AppID,
		WebhookSecret: opts.WebhookSecret,
		PrivateKey:    opts.PrivateKey,
	}
}
func newInstall(installID int64, app *GithubApp) Install {
	return Install{
		ID:        internal.NewID("ghain"),
		InstallID: installID,
		GithubApp: app,
	}
}
