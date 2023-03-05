package module

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/leg100/otf/cloud"
	"github.com/leg100/otf/rbac"
	"github.com/leg100/surl"
)

type (
	// service is the service service for modules
	service interface {
		// PublishModule publishes a module from a VCS repository.
		PublishModule(context.Context, PublishModuleOptions) (*Module, error)
		// CreateModule creates a module without a connection to a VCS repository.
		CreateModule(context.Context, CreateModuleOptions) (*Module, error)
		UpdateModuleStatus(ctx context.Context, opts UpdateModuleStatusOptions) (*Module, error)
		ListModules(context.Context, ListModulesOptions) ([]*Module, error)
		GetModule(ctx context.Context, opts GetModuleOptions) (*Module, error)
		GetModuleByID(ctx context.Context, id string) (*Module, error)
		GetModuleByWebhookID(ctx context.Context, id uuid.UUID) (*Module, error)
		DeleteModule(ctx context.Context, id string) (*Module, error)

		CreateModuleVersion(context.Context, CreateModuleVersionOptions) (*ModuleVersion, error)
		UploadModuleVersion(ctx context.Context, opts UploadModuleVersionOptions) (*ModuleVersion, error)
		DownloadModuleVersion(ctx context.Context, opts DownloadModuleOptions) ([]byte, error)
	}

	Service struct {
		otf.VCSProviderService // for retrieving vcs client
		logr.Logger
		*Publisher

		db *pgdb

		organization otf.Authorizer

		api *api
		web *web
	}
	Options struct {
		OrganizationAuthorizer otf.Authorizer
		CloudService           cloud.Service
		otf.DB
		*surl.Signer
		otf.Renderer
		logr.Logger
	}
)

func NewService(opts Options) *Service {
	svc := Service{
		Logger:       opts.Logger,
		organization: opts.OrganizationAuthorizer,
		db:         &pgdb{opts.DB},
	}

	svc.api = &api{
		svc:    &svc,
		Signer: opts.Signer,
	}
	svc.web = &web{
		Renderer: opts.Renderer,
		app:      &svc,
	}
	return &svc
}

func (s *Service) AddHandlers(r *mux.Router) {
	s.api.addHandlers(r)
	s.web.addHandlers(r)
}

func (a *Service) PublishModule(ctx context.Context, opts PublishModuleOptions) (*Module, error) {
	subject, err := a.organization.CanAccess(ctx, rbac.CreateModuleAction, opts.Organization)
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

func (a *Service) CreateModule(ctx context.Context, opts CreateModuleOptions) (*Module, error) {
	subject, err := a.organization.CanAccess(ctx, rbac.CreateModuleAction, opts.Organization)
	if err != nil {
		return nil, err
	}

	module := newModule(opts)

	if err := a.db.CreateModule(ctx, module); err != nil {
		a.Error(err, "creating module", "subject", subject, "module", module)
		return nil, err
	}
	a.V(0).Info("created module", "subject", subject, "module", module)
	return module, nil
}

func (a *Service) UpdateModuleStatus(ctx context.Context, opts UpdateModuleStatusOptions) (*Module, error) {
	// retrieve module first in order to get organization for authorization
	module, err := a.db.GetModuleByID(ctx, opts.ID)
	if err != nil {
		return nil, err
	}
	organization := module.Organization()

	subject, err := a.organization.CanAccess(ctx, rbac.UpdateModuleAction, organization)
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

func (a *Service) ListModules(ctx context.Context, opts ListModulesOptions) ([]*Module, error) {
	subject, err := a.organization.CanAccess(ctx, rbac.ListModulesAction, opts.Organization)
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

func (a *Service) GetModule(ctx context.Context, opts GetModuleOptions) (*Module, error) {
	subject, err := a.organization.CanAccess(ctx, rbac.GetModuleAction, opts.Organization)
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

func (a *Service) GetModuleByID(ctx context.Context, id string) (*Module, error) {
	module, err := a.db.GetModuleByID(ctx, id)
	if err != nil {
		a.Error(err, "retrieving module", "id", id)
		return nil, err
	}

	subject, err := a.organization.CanAccess(ctx, rbac.GetModuleAction, module.Organization())
	if err != nil {
		return nil, err
	}

	a.V(2).Info("retrieved module", "subject", subject, "module", module)
	return module, nil
}

func (a *Service) GetModuleByWebhookID(ctx context.Context, id uuid.UUID) (*Module, error) {
	return a.db.GetModuleByWebhookID(ctx, id)
}

func (a *Service) DeleteModule(ctx context.Context, id string) (*Module, error) {
	module, err := a.db.GetModuleByID(ctx, id)
	if err != nil {
		a.Error(err, "retrieving module", "id", id)
		return nil, err
	}

	subject, err := a.organization.CanAccess(ctx, rbac.DeleteModuleAction, module.Organization())
	if err != nil {
		return nil, err
	}

	if module.Repo() == nil {
		// there is no webhook to unhook from, so just delete the module
		if err = a.db.DeleteModule(ctx, id); err != nil {
			a.Error(err, "deleting module", "subject", subject, "module", module)
			return nil, err
		}
		a.V(2).Info("deleted module", "subject", subject, "module", module)
		return module, nil
	}

	client, err := a.GetVCSClient(ctx, module.Repo().ProviderID)
	if err != nil {
		return nil, err
	}

	// delete webhook as well as module
	err = a.Unhook(ctx, otf.DisconnectOptions{
		HookID: module.Repo().WebhookID,
		Client: client,
		UnhookCallback: func(ctx context.Context, tx otf.DB) error {
			return deleteModule(ctx, tx, id)
		},
	})
	if err != nil {
		a.Error(err, "deleting module", "subject", subject, "module", module)
		return nil, err
	}

	a.V(2).Info("deleted module", "subject", subject, "module", module)
	return module, nil
}

func (a *Service) CreateModuleVersion(ctx context.Context, opts CreateModuleVersionOptions) (*ModuleVersion, error) {
	module, err := a.db.GetModuleByID(ctx, opts.ModuleID)
	if err != nil {
		return nil, err
	}

	subject, err := a.organization.CanAccess(ctx, rbac.CreateModuleAction, module.organization)
	if err != nil {
		return nil, err
	}

	version, err := module.addNewVersion(opts)

	if err := a.db.CreateModuleVersion(ctx, version); err != nil {
		a.Error(err, "creating module version", "organization", module.organization, "subject", subject, "module_version", version)
		return nil, err
	}
	a.V(0).Info("created module version", "organization", module.organization, "subject", subject, "module_version", version)
	return version, nil
}

func (a *Service) UploadModuleVersion(ctx context.Context, opts UploadModuleVersionOptions) (*ModuleVersion, error) {
	module, err := a.db.GetModuleByID(ctx, opts.ModuleID)
	if err != nil {
		return nil, err
	}

	err = module.upload(opts.Version, opts.TarballGetter)
	if err != nil {
		a.Error(err, "uploading module version", "module_version_id", opts.ModuleVersionID)
		return nil, err
	}
	if modver.Status() != ModuleVersionStatusOk {
		a.Error(err, "uploading module version", "module_version", modver)
		return modver, err
	}

	// check tarball is legit and if not set bad status
	if _, err := UnmarshalTerraformModule(opts.Tarball); err != nil {
		return a.UpdateModuleVersionStatus(ctx, UpdateModuleVersionStatusOptions{
			ID:     opts.ModuleVersionID,
			Status: ModuleVersionStatusRegIngressFailed,
			Error:  err.Error(),
		})
	}

	a.V(0).Info("uploaded module version", "module_version", modver)
	return modver, nil
}

// DownloadModuleVersion should be accessed via signed URL
func (a *Service) DownloadModuleVersion(ctx context.Context, opts DownloadModuleOptions) ([]byte, error) {
	tarball, err := a.db.DownloadModuleVersion(ctx, opts)
	if err != nil {
		a.Error(err, "downloading module", "module_version_id", opts.ModuleVersionID)
		return nil, err
	}
	a.V(2).Info("downloaded module", "module_version_id", opts.ModuleVersionID)
	return tarball, nil
}
