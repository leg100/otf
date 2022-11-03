package app

import (
	"context"

	"github.com/leg100/otf"
)

func (a *Application) CreateVCSProvider(ctx context.Context, opts otf.VCSProviderCreateOptions) (*otf.VCSProvider, error) {
	subject, err := a.CanAccessOrganization(ctx, otf.CreateVCSProviderAction, opts.OrganizationName)
	if err != nil {
		return nil, err
	}

	provider := otf.NewVCSProvider(opts)
	if err := a.db.CreateVCSProvider(ctx, provider); err != nil {
		a.Error(err, "creating vcs provider", "organization", opts.OrganizationName, "id", provider.ID(), "subject", subject)
		return nil, err
	}
	a.V(0).Info("created vcs provider", "organization", opts.OrganizationName, "id", provider.ID(), "subject", subject)
	return provider, nil
}

func (a *Application) ListVCSProviders(ctx context.Context, organization string) ([]*otf.VCSProvider, error) {
	subject, err := a.CanAccessOrganization(ctx, otf.ListVCSProvidersAction, organization)
	if err != nil {
		return nil, err
	}

	tokens, err := a.db.ListVCSProviders(ctx, organization)
	if err != nil {
		a.Error(err, "listing vcs providers", "organization", organization, "subject", subject)
		return nil, err
	}
	a.V(2).Info("listed vcs providers", "organization", organization, "subject", subject)
	return tokens, nil
}

func (a *Application) DeleteVCSProvider(ctx context.Context, id, organization string) error {
	subject, err := a.CanAccessOrganization(ctx, otf.DeleteVCSProviderAction, organization)
	if err != nil {
		return err
	}

	if err := a.db.DeleteVCSProvider(ctx, id); err != nil {
		a.Error(err, "deleting vcs provider", "id", id, "subject", subject)
		return err
	}
	a.V(0).Info("deleted vcs provider", "id", id, "subject", subject)
	return nil
}
