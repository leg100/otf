package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/leg100/otf"
)

func (a *Application) CreateWebhook(ctx context.Context, opts otf.CreateWebhookOptions) (*otf.Webhook, error) {
	unsynced, err := otf.NewUnsynchronisedWebhook(otf.NewUnsynchronisedWebhookOptions{
		Identifier: opts.Identifier,
		Cloud:      opts.Cloud,
	})
	if err != nil {
		a.Error(err, "constructing webhook", "repo", opts.Identifier)
		return nil, err
	}
	hook, err := a.Synchronise(ctx, opts.ProviderID, unsynced)
	if err != nil {
		a.Error(err, "creating webhook", "repo", opts.Identifier)
		return nil, err
	}

	a.V(0).Info("created webhook", "id", hook.ID(), "repo", hook.Identifier())

	return hook, nil
}

func (a *Application) GetWebhook(ctx context.Context, webhookID uuid.UUID) (*otf.Webhook, error) {
	return a.db.GetWebhook(ctx, webhookID)
}

func (a *Application) DeleteWebhook(ctx context.Context, providerID string, webhookID uuid.UUID) error {
	return a.Delete(ctx, providerID, webhookID)
}
