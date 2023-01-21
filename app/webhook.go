package app

import (
	"context"

	"github.com/leg100/otf"
)

func (a *Application) CreateWebhook(ctx context.Context, opts otf.SynchroniseWebhookOptions) (*otf.Webhook, error) {
	hook, err := a.Synchronise(ctx, opts)
	if err != nil {
		a.Error(err, "creating webhook", "repo", opts.Identifier)
		return nil, err
	}

	a.V(0).Info("created webhook", "id", hook.ID, "repo", hook.Identifier())

	return hook, nil
}
