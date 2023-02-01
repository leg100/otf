package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/leg100/otf"
	"github.com/leg100/otf/rbac"
)

func (a *Application) PublishModule(ctx context.Context, opts otf.PublishModuleOptions) (*otf.Module, error) {
	subject, err := a.CanAccessOrganization(ctx, rbac.CreateModuleAction, opts.Organization.Name())
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
	subject, err := a.CanAccessOrganization(ctx, rbac.CreateModuleAction, opts.Organization)
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

func (a *Application) UpdateModuleStatus(ctx context.Context, opts otf.UpdateModuleStatusOptions) (*otf.Module, error) {
	// retrieve module first in order to get organization for authorization
	module, err := a.db.GetModuleByID(ctx, opts.ID)
	if err != nil {
		return nil, err
	}
	organization := module.Organization()

	subject, err := a.CanAccessOrganization(ctx, rbac.UpdateModuleAction, organization)
	if err != nil {
		return nil, err
	}

	module.UpdateStatus(opts.Status)

	if err := a.db.UpdateModuleStatus(ctx, opts); err != nil {
		a.Error(err, "updating module status", "subject", subject, "module", module, "status", opts.Status)
		return nil, err
	}
	a.V(0).Info("updated module status", "subject", subject, "module", module, "status", opts.Status)
	return module, nil
}

func (a *Application) ListModules(ctx context.Context, opts otf.ListModulesOptions) ([]*otf.Module, error) {
	subject, err := a.CanAccessOrganization(ctx, rbac.ListModulesAction, opts.Organization)
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
	subject, err := a.CanAccessOrganization(ctx, rbac.GetModuleAction, opts.Organization)
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

	subject, err := a.CanAccessOrganization(ctx, rbac.GetModuleAction, module.Organization())
	if err != nil {
		return nil, err
	}

	a.V(2).Info("retrieved module", "subject", subject, "module", module)
	return module, nil
}

func (a *Application) GetModuleByWebhookID(ctx context.Context, id uuid.UUID) (*otf.Module, error) {
	return a.db.GetModuleByWebhookID(ctx, id)
}

func (a *Application) DeleteModule(ctx context.Context, id string) (*otf.Module, error) {
	module, err := a.db.GetModuleByID(ctx, id)
	if err != nil {
		a.Error(err, "retrieving module", "id", id)
		return nil, err
	}

	subject, err := a.CanAccessOrganization(ctx, rbac.DeleteModuleAction, module.Organization())
	if err != nil {
		return nil, err
	}

	// TODO: delete webhook

	if err = a.db.DeleteModule(ctx, id); err != nil {
		a.Error(err, "deleting module", "subject", subject, "module", module)
		return nil, err
	}

	a.V(2).Info("deleted module", "subject", subject, "module", module)
	return module, nil
}
