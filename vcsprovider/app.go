package vcsprovider

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
	"github.com/leg100/otf/cloud"
	"github.com/leg100/otf/rbac"
)

type service interface {
	create(ctx context.Context, opts createOptions) (*VCSProvider, error)
	get(ctx context.Context, id string) (*VCSProvider, error)
	list(ctx context.Context, organization string) ([]*VCSProvider, error)
	delete(ctx context.Context, id string) (*VCSProvider, error)

	// GetVCSClient combines retrieving a vcs provider and construct a cloud
	// client from that provider.
	//
	// TODO: rename vcs provider to cloud client; the central purpose of the vcs
	// provider is, after all, to construct a cloud client.
	GetVCSClient(ctx context.Context, providerID string) (cloud.Client, error)
}

// Application is the implementation of appService
type Application struct {
	otf.Authorizer
	logr.Logger

	db
	factory
}

func (a *Application) create(ctx context.Context, opts createOptions) (*VCSProvider, error) {
	subject, err := a.CanAccessOrganization(ctx, rbac.CreateVCSProviderAction, opts.Organization)
	if err != nil {
		return nil, err
	}

	provider, err := a.new(opts)
	if err != nil {
		return nil, err
	}

	if err := a.db.create(ctx, provider); err != nil {
		a.Error(err, "creating vcs provider", "organization", opts.Organization, "id", provider.ID(), "subject", subject)
		return nil, err
	}
	a.V(0).Info("created vcs provider", "organization", opts.Organization, "id", provider.ID(), "subject", subject)
	return provider, nil
}

func (a *Application) get(ctx context.Context, id string) (*VCSProvider, error) {
	// Parameters only include VCS Provider ID, so we can only determine
	// authorization _after_ retrieving the provider

	provider, err := a.db.get(ctx, id)
	if err != nil {
		a.Error(err, "retrieving vcs provider", "id", id)
		return nil, err
	}

	subject, err := a.CanAccessOrganization(ctx, rbac.GetVCSProviderAction, provider.Organization())
	if err != nil {
		return nil, err
	}
	a.V(2).Info("retrieved vcs provider", "provider", provider, "subject", subject)

	return provider, nil
}

func (a *Application) list(ctx context.Context, organization string) ([]*VCSProvider, error) {
	subject, err := a.CanAccessOrganization(ctx, rbac.ListVCSProvidersAction, organization)
	if err != nil {
		return nil, err
	}

	providers, err := a.db.list(ctx, organization)
	if err != nil {
		a.Error(err, "listing vcs providers", "organization", organization, "subject", subject)
		return nil, err
	}
	a.V(2).Info("listed vcs providers", "organization", organization, "subject", subject)
	return providers, nil
}

func (a *Application) delete(ctx context.Context, id string) (*VCSProvider, error) {
	// retrieve vcs provider first in order to get organization for authorization
	provider, err := a.db.get(ctx, id)
	if err != nil {
		a.Error(err, "retrieving vcs provider", "id", id)
		return nil, err
	}

	subject, err := a.CanAccessOrganization(ctx, rbac.DeleteVCSProviderAction, provider.Organization())
	if err != nil {
		return nil, err
	}

	if err := a.db.delete(ctx, id); err != nil {
		a.Error(err, "deleting vcs provider", "provider", provider, "subject", subject)
		return nil, err
	}
	a.V(0).Info("deleted vcs provider", "provider", provider, "subject", subject)
	return provider, nil
}

func (a *Application) GetVCSClient(ctx context.Context, providerID string) (cloud.Client, error) {
	provider, err := a.get(ctx, providerID)
	if err != nil {
		return nil, err
	}
	return provider.NewClient(ctx)
}
