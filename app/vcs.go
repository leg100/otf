package app

import (
	"context"

	"github.com/leg100/otf"
)

func (a *Application) SetStatus(ctx context.Context, providerID string, opts otf.SetStatusOptions) error {
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

func (a *Application) GetRepoTarball(ctx context.Context, providerID string, opts otf.GetRepoTarballOptions) ([]byte, error) {
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

func (a *Application) GetRepository(ctx context.Context, providerID string, identifier string) (*otf.Repo, error) {
	provider, err := a.db.GetVCSProvider(ctx, providerID)
	if err != nil {
		return nil, err
	}
	client, err := provider.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	return client.GetRepository(ctx, identifier)
}

func (a *Application) ListRepositories(ctx context.Context, providerID string, opts otf.ListOptions) (*otf.RepoList, error) {
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

func (a *Application) CreateWebhook(ctx context.Context, providerID string, opts otf.CreateCloudWebhookOptions) error {
	provider, err := a.db.GetVCSProvider(ctx, providerID)
	if err != nil {
		return err
	}
	client, err := provider.NewClient(ctx)
	if err != nil {
		return err
	}
	return client.CreateWebhook(ctx, opts)
}
