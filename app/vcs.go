package app

import (
	"context"

	"github.com/leg100/otf/cloud"
)

func (a *Application) SetStatus(ctx context.Context, providerID string, opts cloud.SetStatusOptions) error {
	provider, err := a.db.GetVCSProvider(ctx, providerID)
	if err != nil {
		return err
	}
	client, err := provider.NewClient(ctx)
	if err != nil {
		return err
	}
	return client.SetStatus(ctx, opts)
}

func (a *Application) GetRepoTarball(ctx context.Context, providerID string, opts cloud.GetRepoTarballOptions) ([]byte, error) {
	provider, err := a.db.GetVCSProvider(ctx, providerID)
	if err != nil {
		return nil, err
	}
	client, err := provider.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	return client.GetRepoTarball(ctx, opts)
}

func (a *Application) GetRepository(ctx context.Context, providerID string, identifier string) (cloud.Repo, error) {
	provider, err := a.db.GetVCSProvider(ctx, providerID)
	if err != nil {
		return cloud.Repo{}, err
	}
	client, err := provider.NewClient(ctx)
	if err != nil {
		return cloud.Repo{}, err
	}
	return client.GetRepository(ctx, identifier)
}

func (a *Application) ListRepositories(ctx context.Context, providerID string, opts cloud.ListRepositoriesOptions) ([]cloud.Repo, error) {
	provider, err := a.db.GetVCSProvider(ctx, providerID)
	if err != nil {
		return nil, err
	}
	client, err := provider.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	return client.ListRepositories(ctx, opts)
}

func (a *Application) ListTags(ctx context.Context, providerID string, opts cloud.ListTagsOptions) ([]string, error) {
	provider, err := a.db.GetVCSProvider(ctx, providerID)
	if err != nil {
		return nil, err
	}
	client, err := provider.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	return client.ListTags(ctx, opts)
}

func (a *Application) CreateWebhook(ctx context.Context, providerID string, opts cloud.CreateWebhookOptions) (string, error) {
	provider, err := a.db.GetVCSProvider(ctx, providerID)
	if err != nil {
		return "", err
	}
	client, err := provider.NewClient(ctx)
	if err != nil {
		return "", err
	}
	return client.CreateWebhook(ctx, opts)
}

func (a *Application) UpdateWebhook(ctx context.Context, providerID string, opts cloud.UpdateWebhookOptions) error {
	provider, err := a.db.GetVCSProvider(ctx, providerID)
	if err != nil {
		return err
	}
	client, err := provider.NewClient(ctx)
	if err != nil {
		return err
	}
	return client.UpdateWebhook(ctx, opts)
}

func (a *Application) GetWebhook(ctx context.Context, providerID string, opts cloud.GetWebhookOptions) (cloud.Webhook, error) {
	provider, err := a.db.GetVCSProvider(ctx, providerID)
	if err != nil {
		return cloud.Webhook{}, err
	}
	client, err := provider.NewClient(ctx)
	if err != nil {
		return cloud.Webhook{}, err
	}
	return client.GetWebhook(ctx, opts)
}

func (a *Application) DeleteWebhook(ctx context.Context, providerID string, opts cloud.DeleteWebhookOptions) error {
	provider, err := a.db.GetVCSProvider(ctx, providerID)
	if err != nil {
		return err
	}
	client, err := provider.NewClient(ctx)
	if err != nil {
		return err
	}
	return client.DeleteWebhook(ctx, opts)
}
