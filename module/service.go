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
	// service is the service for modules
	service interface {
		// PublishModule publishes a module from a VCS repository.
		PublishModule(context.Context, otf.PublishModuleOptions) (*otf.Module, error)
		// CreateModule creates a module without a connection to a VCS repository.
		CreateModule(context.Context, otf.CreateModuleOptions) (*otf.Module, error)
		UpdateModuleStatus(ctx context.Context, opts otf.UpdateModuleStatusOptions) (*otf.Module, error)
		ListModules(context.Context, otf.ListModulesOptions) ([]*otf.Module, error)
		GetModule(ctx context.Context, opts otf.GetModuleOptions) (*otf.Module, error)
		GetModuleByID(ctx context.Context, id string) (*otf.Module, error)
		GetModuleByRepoID(ctx context.Context, repoID uuid.UUID) (*otf.Module, error)
		DeleteModule(ctx context.Context, id string) (*otf.Module, error)

		CreateModuleVersion(context.Context, otf.CreateModuleVersionOptions) (*otf.ModuleVersion, error)

		uploadVersion(ctx context.Context, versionID string, tarball []byte) (*otf.ModuleVersion, error)
		downloadVersion(ctx context.Context, versionID string) ([]byte, error)
	}

	Service struct {
		otf.VCSProviderService
		logr.Logger
		*Publisher

		db   *pgdb
		repo otf.RepoService

		organization otf.Authorizer

		api *api
		web *web
	}

	Options struct {
		OrganizationAuthorizer otf.Authorizer
		CloudService           cloud.Service

		otf.DB
		otf.VCSProviderService
		*surl.Signer
		otf.Renderer
		otf.RepoService
		logr.Logger
	}
)

func NewService(opts Options) *Service {
	svc := Service{
		Logger:       opts.Logger,
		organization: opts.OrganizationAuthorizer,
		db:           &pgdb{opts.DB},
		repo:         opts.RepoService,
	}

	svc.api = &api{
		svc:    &svc,
		Signer: opts.Signer,
	}
	svc.web = &web{
		Renderer: opts.Renderer,
		svc:      &svc,
	}
	return &svc
}

func serviceWithDB(parent *Service, db *pgdb) *Service {
	child := *parent
	child.db = db
	return &child
}

func (s *Service) AddHandlers(r *mux.Router) {
	s.api.addHandlers(r)
	s.web.addHandlers(r)
}

func (s *Service) PublishModule(ctx context.Context, opts otf.PublishModuleOptions) (*otf.Module, error) {
	vcsprov, err := s.GetVCSProvider(ctx, opts.VCSProviderID)
	if err != nil {
		return nil, err
	}

	subject, err := s.organization.CanAccess(ctx, rbac.CreateModuleAction, vcsprov.Organization)
	if err != nil {
		return nil, err
	}

	module, err := s.Publisher.PublishModule(ctx, opts)
	if err != nil {
		s.Error(err, "publishing module", "subject", subject, "repo", opts.RepoPath)
		return nil, err
	}
	s.V(0).Info("published module", "subject", subject, "module", module)
	return module, nil
}

func (a *Service) CreateModule(ctx context.Context, opts otf.CreateModuleOptions) (*otf.Module, error) {
	subject, err := a.organization.CanAccess(ctx, rbac.CreateModuleAction, opts.Organization)
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

func (a *Service) UpdateModuleStatus(ctx context.Context, opts otf.UpdateModuleStatusOptions) (*otf.Module, error) {
	module.UpdateStatus(opts.Status)

	if err := a.db.UpdateModuleStatus(ctx, opts); err != nil {
		a.Error(err, "updating module status", "subject", subject, "module", module, "status", opts.Status)
		return nil, err
	}
	a.V(0).Info("updated module status", "subject", subject, "module", module, "status", opts.Status)
	return module, nil
}

func (a *Service) ListModules(ctx context.Context, opts ListModulesOptions) ([]*otf.Module, error) {
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

func (a *Service) GetModule(ctx context.Context, opts GetModuleOptions) (*otf.Module, error) {
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

func (a *Service) GetModuleByID(ctx context.Context, id string) (*otf.Module, error) {
	module, err := a.db.GetModuleByID(ctx, id)
	if err != nil {
		a.Error(err, "retrieving module", "id", id)
		return nil, err
	}

	subject, err := a.organization.CanAccess(ctx, rbac.GetModuleAction, module.Organization)
	if err != nil {
		return nil, err
	}

	a.V(2).Info("retrieved module", "subject", subject, "module", module)
	return module, nil
}

func (a *Service) GetModuleByWebhookID(ctx context.Context, id uuid.UUID) (*otf.Module, error) {
	return a.db.GetModuleByWebhookID(ctx, id)
}

func (a *Service) DeleteModule(ctx context.Context, id string) (*otf.Module, error) {
	module, err := a.db.GetModuleByID(ctx, id)
	if err != nil {
		a.Error(err, "retrieving module", "id", id)
		return nil, err
	}

	subject, err := a.organization.CanAccess(ctx, rbac.DeleteModuleAction, module.Organization)
	if err != nil {
		return nil, err
	}

	err = a.db.tx(ctx, func(tx *pgdb) error {
		// disconnect module prior to deletion
		if module.Connection != nil {
			err := a.repo.Disconnect(ctx, otf.DisconnectOptions{
				ConnectionType: otf.ModuleConnection,
				ResourceID:     module.ID,
				Tx:             tx,
			})
			if err != nil {
				return err
			}
		}
		return tx.DeleteModule(ctx, id)
	})
	if err != nil {
		a.Error(err, "deleting module", "subject", subject, "module", module)
		return nil, err
	}
	a.V(2).Info("deleted module", "subject", subject, "module", module)
	return module, nil
}

func (a *Service) CreateModuleVersion(ctx context.Context, opts otf.CreateModuleVersionOptions) (*otf.ModuleVersion, error) {
	module, err := a.db.GetModuleByID(ctx, opts.ModuleID)
	if err != nil {
		return nil, err
	}

	subject, err := a.organization.CanAccess(ctx, rbac.CreateModuleVersionAction, module.Organization)
	if err != nil {
		return nil, err
	}

	modver := otf.NewModuleVersion(opts)
	if err := module.AddVersion(modver); err != nil {
		return nil, err
	}

	if err := a.db.CreateModuleVersion(ctx, modver); err != nil {
		a.Error(err, "creating module version", "organization", module.Organization, "subject", subject, "module_version", modver)
		return nil, err
	}
	a.V(0).Info("created module version", "organization", module.Organization, "subject", subject, "module_version", modver)
	return modver, nil
}

func (a *Service) uploadVersion(ctx context.Context, versionID string, tarball []byte) error {
	module, err := a.db.getModuleByVersionID(ctx, versionID)
	if err != nil {
		return err
	}

	// validate tarball
	if _, err := unmarshalTerraformModule(tarball); err != nil {
		a.Error(err, "uploading module version", "module_version", versionID)
		_, err = a.db.UpdateModuleVersionStatus(ctx, otf.UpdateModuleVersionStatusOptions{
			ID:     versionID,
			Status: otf.ModuleVersionStatusRegIngressFailed,
			Error:  err.Error(),
		})
		return err
	}

	// save tarball, set status, and make it the latest version
	err = a.db.tx(ctx, func(tx *pgdb) error {
		if err := tx.saveTarball(ctx, versionID, tarball); err != nil {
			return err
		}
		_, err := tx.UpdateModuleVersionStatus(ctx, otf.UpdateModuleVersionStatusOptions{
			ID:     versionID,
			Status: otf.ModuleVersionStatusOK,
		})
		if err != nil {
			return err
		}
		return tx.updateLatest(ctx, module.ID, versionID)
	})
	if err != nil {
		a.Error(err, "uploading module version", "module_version_id", versionID)
		return err
	}

	a.V(0).Info("uploaded module version", "module_version", versionID)
	return nil
}

// downloadVersion should be accessed via signed URL
func (a *Service) downloadVersion(ctx context.Context, versionID string) ([]byte, error) {
	tarball, err := a.db.getTarball(ctx, versionID)
	if err != nil {
		a.Error(err, "downloading module", "module_version_id", versionID)
		return nil, err
	}
	a.V(2).Info("downloaded module", "module_version_id", versionID)
	return tarball, nil
}

func (a *Service) deleteVersion(ctx context.Context, versionID string) (*otf.Module, error) {
	module, err := a.db.GetModuleByID(ctx, versionID)
	if err != nil {
		a.Error(err, "retrieving module", "id", versionID)
		return nil, err
	}

	subject, err := a.organization.CanAccess(ctx, rbac.DeleteModuleVersionAction, module.Organization)
	if err != nil {
		return nil, err
	}

	if err := module.RemoveVersion(versionID); err != nil {
		return nil, err
	}

	err = a.db.tx(ctx, func(tx *pgdb) error {
		if err := tx.deleteModuleVersion(ctx, versionID); err != nil {
			return err
		}
		if module.SetLatest() {
			if err := tx.updateLatest(ctx, module.ID, module.Latest); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		a.Error(err, "deleting module", "subject", subject, "module", module)
		return nil, err
	}
	a.V(2).Info("deleted module", "subject", subject, "module", module)
	return module, nil
}

// tx returns a service in a callback, with its database calls wrapped within a transaction
func (a *Service) tx(ctx context.Context, txFunc func(*Service) error) error {
	return a.db.Tx(ctx, func(db otf.DB) error {
		return txFunc(serviceWithDB(a, &pgdb{db}))
	})
}
