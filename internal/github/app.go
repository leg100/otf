package github

type (
	App struct {
		ID            int64 // github's app id
		WebhookSecret string
		PrivateKey    string
	}

	Install struct {
		GithubID int64 // github's install ID
		App      *App
	}
)

func newApp(opts CreateAppOptions) *App {
	return &App{
		ID:            opts.AppID,
		WebhookSecret: opts.WebhookSecret,
		PrivateKey:    opts.PrivateKey,
	}
}
func newInstall(installID int64, app *App) Install {
	return Install{
		GithubID: installID,
		App:      app,
	}
}
