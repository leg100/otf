package github

type (
	App struct {
		ID            int64 // github's app id
		WebhookSecret string
		PrivateKey    string
	}

	Install struct {
		ID  int64 // github's install ID
		App *App
	}
)

func newApp(opts CreateAppOptions) *App {
	return &App{
		ID:            opts.AppID,
		WebhookSecret: opts.WebhookSecret,
		PrivateKey:    opts.PrivateKey,
	}
}
