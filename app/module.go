package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/leg100/otf"
)

func (a *Application) CreateModule(ctx context.Context, opts otf.CreateModuleOptions) (*otf.Module, error) {
	subject, err := a.CanAccessOrganization(ctx, otf.CreateModuleAction, opts.Organization.Name())
	if err != nil {
		return nil, err
	}

	module := otf.NewModule(opts)

	if err := a.db.CreateModule(ctx, module); err != nil {
		a.Error(err, "creating module", "subject", subject, "module", module)
		return nil, err
	}
	a.V(0).Info("created module", "subject", subject, "module", module)
	return module, nil
}

func (a *Application) CreateModuleVersion(ctx context.Context, opts otf.CreateModuleVersionOptions) (*otf.ModuleVersion, error) {
	// retrieve module first in order to get organization for authorization
	module, err := a.db.GetModuleByID(ctx, opts.ModuleID)
	if err != nil {
		return nil, err
	}
	organization := module.Organization().Name()

	subject, err := a.CanAccessOrganization(ctx, otf.CreateModuleAction, organization)
	if err != nil {
		return nil, err
	}

	version := otf.NewModuleVersion(opts)

	if err := a.db.CreateModuleVersion(ctx, version); err != nil {
		a.Error(err, "creating module version", "organization", organization, "subject", subject, "module_version", version)
		return nil, err
	}
	a.V(0).Info("created module version", "organization", organization, "subject", subject, "module_version", version)
	return version, nil
}

func (a *Application) ListModules(ctx context.Context, opts otf.ListModulesOptions) ([]*otf.Module, error) {
	subject, err := a.CanAccessOrganization(ctx, otf.ListModulesAction, opts.Organization)
	if err != nil {
		return nil, err
	}

	modules, err := a.db.ListModules(ctx, opts)
	if err != nil {
		a.Error(err, "listing modules", "organization", opts.Organization, "subject", subject)
		return nil, err
	}
	a.V(2).Info("listed modules", "organization", opts.Organization, "subject", subject)
	return modules, nil
}

func (a *Application) GetModule(ctx context.Context, opts otf.GetModuleOptions) (*otf.Module, error) {
	subject, err := a.CanAccessOrganization(ctx, otf.ListModulesAction, opts.Organization)
	if err != nil {
		return nil, err
	}
	module, err := a.db.GetModule(ctx, opts)
	if err != nil {
		a.Error(err, "retrieving module", "organization", opts.Organization, "subject", subject, "name", opts.Name, "provider", opts.Provider)
		return nil, err
	}
	a.V(2).Info("listed modules", "subject", subject, "module", module)
	return module, nil
}

func (a *Application) GetModuleByWebhookID(ctx context.Context, id uuid.UUID) (*otf.Module, error) {
	return a.db.GetModuleByWebhookID(ctx, id)
}

func (a *Application) UploadModuleVersion(ctx context.Context, opts otf.UploadModuleVersionOptions) error {
	return a.db.UploadModuleVersion(ctx, opts)
}

func (a *Application) DownloadModuleVersion(ctx context.Context, opts otf.DownloadModuleOptions) ([]byte, error) {
	return a.db.DownloadModuleVersion(ctx, opts)
}
