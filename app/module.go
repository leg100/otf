package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/leg100/otf"
)

func (a *Application) PublishModule(ctx context.Context, opts otf.PublishModuleOptions) (*otf.Module, error) {
	subject, err := a.CanAccessOrganization(ctx, otf.CreateModuleAction, opts.Organization.Name())
	if err != nil {
		return nil, err
	}

	module, err := a.Publisher.PublishModule(ctx, opts)
	if err != nil {
		a.Error(err, "publishing module", "subject", subject, "repo", opts.Identifier)
		return nil, err
	}
	a.V(0).Info("published module", "subject", subject, "module", module)
	return module, nil
}

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
	subject, err := a.CanAccessOrganization(ctx, otf.GetModuleAction, opts.Organization)
	if err != nil {
		return nil, err
	}

	module, err := a.db.GetModule(ctx, opts)
	if err != nil {
		a.Error(err, "retrieving module", "module", opts)
		return nil, err
	}

	a.V(2).Info("retrieved module", "subject", subject, "module", module)
	return module, nil
}

func (a *Application) GetModuleByID(ctx context.Context, id string) (*otf.Module, error) {
	module, err := a.db.GetModuleByID(ctx, id)
	if err != nil {
		a.Error(err, "retrieving module", "id", id)
		return nil, err
	}

	subject, err := a.CanAccessOrganization(ctx, otf.GetModuleAction, module.Organization().Name())
	if err != nil {
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

// DownloadModuleVersion should be accessed via signed URL
func (a *Application) DownloadModuleVersion(ctx context.Context, opts otf.DownloadModuleOptions) ([]byte, error) {
	tarball, err := a.db.DownloadModuleVersion(ctx, opts)
	if err != nil {
		a.Error(err, "downloading module", "module_version_id", opts.ModuleVersionID)
		return nil, err
	}
	a.V(2).Info("downloaded module", "module_version_id", opts.ModuleVersionID)
	return tarball, nil
}

func (a *Application) DeleteModule(ctx context.Context, id string) error {
	module, err := a.db.GetModuleByID(ctx, id)
	if err != nil {
		a.Error(err, "retrieving module", "id", id)
		return err
	}

	subject, err := a.CanAccessOrganization(ctx, otf.DeleteModuleAction, module.Organization().Name())
	if err != nil {
		return err
	}

	err = a.db.DeleteModule(ctx, id)
	if err != nil {
		a.Error(err, "deleting module", "subject", subject, "module", module)
		return err
	}

	a.V(2).Info("deleted module", "subject", subject, "module", module)
	return nil
}
