package module

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/leg100/otf/cloud"
	"github.com/leg100/otf/organization"
	"github.com/leg100/otf/rbac"
	"github.com/leg100/otf/repo"
	"github.com/leg100/otf/semver"
	"github.com/leg100/otf/vcsprovider"
	"github.com/leg100/surl"
)

type (
	ModuleService = Service

	Service interface {
		// PublishModule publishes a module from a VCS repository.
		PublishModule(context.Context, PublishOptions) (*Module, error)
		PublishVersion(context.Context, PublishVersionOptions) error
		// CreateModule creates a module without a connection to a VCS repository.
		CreateModule(context.Context, CreateOptions) (*Module, error)
		ListModules(context.Context, ListModulesOptions) ([]*Module, error)
		GetModule(ctx context.Context, opts GetModuleOptions) (*Module, error)
		GetModuleByID(ctx context.Context, id string) (*Module, error)
		GetModuleByRepoID(ctx context.Context, repoID uuid.UUID) (*Module, error)
		DeleteModule(ctx context.Context, id string) (*Module, error)
		GetModuleInfo(ctx context.Context, versionID string) (*TerraformModule, error)

		CreateVersion(context.Context, CreateModuleVersionOptions) (*ModuleVersion, error)

		uploadVersion(ctx context.Context, versionID string, tarball []byte) error
		downloadVersion(ctx context.Context, versionID string) ([]byte, error)

		updateModuleStatus(ctx context.Context, module *Module, status ModuleStatus) (*Module, error)
	}

	service struct {
		vcsprovider.VCSProviderService
		logr.Logger

		db   *pgdb
		repo repo.Service

		organization otf.Authorizer

		api *api
		web *webHandlers
	}

	Options struct {
		logr.Logger

		otf.DB
		otf.HostnameService
		vcsprovider.VCSProviderService
		*surl.Signer
		otf.Renderer
		repo.RepoService
	}
)

func NewService(opts Options) *service {
	svc := service{
		Logger:             opts.Logger,
		VCSProviderService: opts.VCSProviderService,
		organization:       &organization.Authorizer{opts.Logger},
		db:                 &pgdb{opts.DB},
		repo:               opts.RepoService,
	}

	svc.api = &api{
		svc:    &svc,
		Signer: opts.Signer,
	}
	svc.web = &webHandlers{
		HostnameService:    opts.HostnameService,
		Renderer:           opts.Renderer,
		VCSProviderService: opts.VCSProviderService,
		svc:                &svc,
	}
	return &svc
}

func (s *service) AddHandlers(r *mux.Router) {
	s.api.addHandlers(r)
	s.web.addHandlers(r)
}

// PublishModule publishes a new module from a VCS repository, enumerating through
// its git tags and releasing a module version for each tag.
func (s *service) PublishModule(ctx context.Context, opts PublishOptions) (*Module, error) {
	vcsprov, err := s.GetVCSProvider(ctx, opts.VCSProviderID)
	if err != nil {
		return nil, err
	}

	subject, err := s.organization.CanAccess(ctx, rbac.CreateModuleAction, vcsprov.Organization)
	if err != nil {
		return nil, err
	}

	module, err := s.publishModule(ctx, vcsprov.Organization, opts)
	if err != nil {
		s.Error(err, "publishing module", "subject", subject, "repo", opts.Repo)
		return nil, err
	}
	s.V(0).Info("published module", "subject", subject, "module", module)

	return module, nil
}

func (s *service) publishModule(ctx context.Context, organization string, opts PublishOptions) (*Module, error) {

	name, provider, err := opts.Repo.Split()
	if err != nil {
		return nil, err
	}

	mod := NewModule(CreateOptions{
		Name:         name,
		Provider:     provider,
		Organization: organization,
	})

	// persist module to db and connect to repository
	err = s.db.tx(ctx, func(tx *pgdb) error {
		if err := tx.createModule(ctx, mod); err != nil {
			return err
		}
		connection, err := s.repo.Connect(ctx, repo.ConnectOptions{
			ConnectionType: repo.ModuleConnection,
			ResourceID:     mod.ID,
			VCSProviderID:  opts.VCSProviderID,
			RepoPath:       string(opts.Repo),
			Tx:             tx,
		})
		if err != nil {
			return err
		}
		mod.Connection = connection
		return nil
	})
	if err != nil {
		return nil, err
	}

	client, err := s.GetVCSClient(ctx, opts.VCSProviderID)
	if err != nil {
		return nil, err
	}

	// Make new version for each tag that looks like a semantic version.
	tags, err := client.ListTags(ctx, cloud.ListTagsOptions{
		Repo: string(opts.Repo),
	})
	if err != nil {
		return nil, err
	}
	if len(tags) == 0 {
		return s.updateModuleStatus(ctx, mod, ModuleStatusNoVersionTags)
	}
	for _, tag := range tags {
		// tags/<version> -> <version>
		_, version, found := strings.Cut(tag, "/")
		if !found {
			return nil, fmt.Errorf("malformed git ref: %s", tag)
		}
		// skip tags that are not semantic versions
		if !semver.IsValid(version) {
			continue
		}
		err := s.PublishVersion(ctx, PublishVersionOptions{
			ModuleID: mod.ID,
			// strip off v prefix if it has one
			Version: strings.TrimPrefix(version, "v"),
			Ref:     tag,
			Repo:    opts.Repo,
			Client:  client,
		})
		if err != nil {
			return nil, err
		}
	}

	// return module from db complete with versions
	return s.GetModuleByID(ctx, mod.ID)
}

// PublishVersion publishes a module version, retrieving its contents from a repository and
// uploading it to the module store.
func (s *service) PublishVersion(ctx context.Context, opts PublishVersionOptions) error {
	modver, err := s.CreateVersion(ctx, CreateModuleVersionOptions{
		ModuleID: opts.ModuleID,
		Version:  opts.Version,
	})
	if err != nil {
		return err
	}

	tarball, err := opts.Client.GetRepoTarball(ctx, cloud.GetRepoTarballOptions{
		Repo: string(opts.Repo),
		Ref:  &opts.Ref,
	})
	if err != nil {
		return s.db.updateModuleVersionStatus(ctx, UpdateModuleVersionStatusOptions{
			ID:     modver.ID,
			Status: ModuleVersionStatusCloneFailed,
			Error:  err.Error(),
		})
	}

	return s.uploadVersion(ctx, modver.ID, tarball)
}

func (s *service) CreateModule(ctx context.Context, opts CreateOptions) (*Module, error) {
	subject, err := s.organization.CanAccess(ctx, rbac.CreateModuleAction, opts.Organization)
	if err != nil {
		return nil, err
	}

	module := NewModule(opts)

	if err := s.db.createModule(ctx, module); err != nil {
		s.Error(err, "creating module", "subject", subject, "module", module)
		return nil, err
	}
	s.V(0).Info("created module", "subject", subject, "module", module)
	return module, nil
}

func (s *service) ListModules(ctx context.Context, opts ListModulesOptions) ([]*Module, error) {
	subject, err := s.organization.CanAccess(ctx, rbac.ListModulesAction, opts.Organization)
	if err != nil {
		return nil, err
	}

	modules, err := s.db.listModules(ctx, opts)
	if err != nil {
		s.Error(err, "listing modules", "organization", opts.Organization, "subject", subject)
		return nil, err
	}
	s.V(2).Info("listed modules", "organization", opts.Organization, "subject", subject)
	return modules, nil
}

func (s *service) GetModule(ctx context.Context, opts GetModuleOptions) (*Module, error) {
	subject, err := s.organization.CanAccess(ctx, rbac.GetModuleAction, opts.Organization)
	if err != nil {
		return nil, err
	}

	module, err := s.db.getModule(ctx, opts)
	if err != nil {
		s.Error(err, "retrieving module", "module", opts)
		return nil, err
	}

	s.V(2).Info("retrieved module", "subject", subject, "module", module)
	return module, nil
}

func (s *service) GetModuleByID(ctx context.Context, id string) (*Module, error) {
	module, err := s.db.getModuleByID(ctx, id)
	if err != nil {
		s.Error(err, "retrieving module", "id", id)
		return nil, err
	}

	subject, err := s.organization.CanAccess(ctx, rbac.GetModuleAction, module.Organization)
	if err != nil {
		return nil, err
	}

	s.V(2).Info("retrieved module", "subject", subject, "module", module)
	return module, nil
}

func (s *service) GetModuleByRepoID(ctx context.Context, id uuid.UUID) (*Module, error) {
	return s.db.getModuleByWebhookID(ctx, id)
}

func (s *service) DeleteModule(ctx context.Context, id string) (*Module, error) {
	module, err := s.db.getModuleByID(ctx, id)
	if err != nil {
		s.Error(err, "retrieving module", "id", id)
		return nil, err
	}

	subject, err := s.organization.CanAccess(ctx, rbac.DeleteModuleAction, module.Organization)
	if err != nil {
		return nil, err
	}

	err = s.db.tx(ctx, func(tx *pgdb) error {
		// disconnect module prior to deletion
		if module.Connection != nil {
			err := s.repo.Disconnect(ctx, repo.DisconnectOptions{
				ConnectionType: repo.ModuleConnection,
				ResourceID:     module.ID,
				Tx:             tx,
			})
			if err != nil {
				return err
			}
		}
		return tx.delete(ctx, id)
	})
	if err != nil {
		s.Error(err, "deleting module", "subject", subject, "module", module)
		return nil, err
	}
	s.V(2).Info("deleted module", "subject", subject, "module", module)
	return module, nil
}

func (s *service) CreateVersion(ctx context.Context, opts CreateModuleVersionOptions) (*ModuleVersion, error) {
	module, err := s.db.getModuleByID(ctx, opts.ModuleID)
	if err != nil {
		return nil, err
	}

	subject, err := s.organization.CanAccess(ctx, rbac.CreateModuleVersionAction, module.Organization)
	if err != nil {
		return nil, err
	}

	modver := NewModuleVersion(opts)

	if err := s.db.createModuleVersion(ctx, modver); err != nil {
		s.Error(err, "creating module version", "organization", module.Organization, "subject", subject, "module_version", modver)
		return nil, err
	}
	s.V(0).Info("created module version", "organization", module.Organization, "subject", subject, "module_version", modver)
	return modver, nil
}

func (s *service) GetModuleInfo(ctx context.Context, versionID string) (*TerraformModule, error) {
	tarball, err := s.db.getTarball(ctx, versionID)
	if err != nil {
		return nil, err
	}
	return unmarshalTerraformModule(tarball)
}

func (s *service) updateModuleStatus(ctx context.Context, mod *Module, status ModuleStatus) (*Module, error) {
	mod.Status = status

	if err := s.db.updateModuleStatus(ctx, mod.ID, status); err != nil {
		s.Error(err, "updating module status", "module", mod.ID, "status", status)
		return nil, err
	}
	s.V(0).Info("updated module status", "module", mod.ID, "status", status)
	return mod, nil
}

func (s *service) uploadVersion(ctx context.Context, versionID string, tarball []byte) error {
	module, err := s.db.getModuleByVersionID(ctx, versionID)
	if err != nil {
		return err
	}

	// validate tarball
	if _, err := unmarshalTerraformModule(tarball); err != nil {
		s.Error(err, "uploading module version", "module_version", versionID)
		return s.db.updateModuleVersionStatus(ctx, UpdateModuleVersionStatusOptions{
			ID:     versionID,
			Status: ModuleVersionStatusRegIngressFailed,
			Error:  err.Error(),
		})
	}

	// save tarball, set status, and make it the latest version
	err = s.db.tx(ctx, func(tx *pgdb) error {
		if err := tx.saveTarball(ctx, versionID, tarball); err != nil {
			return err
		}
		err = tx.updateModuleVersionStatus(ctx, UpdateModuleVersionStatusOptions{
			ID:     versionID,
			Status: ModuleVersionStatusOK,
		})
		if err != nil {
			return err
		}
		// re-retrieve module so that includes the above version with updated
		// status
		module, err = tx.getModuleByID(ctx, module.ID)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		s.Error(err, "uploading module version", "module_version_id", versionID)
		return err
	}

	s.V(0).Info("uploaded module version", "module_version", versionID)
	return nil
}

// downloadVersion should be accessed via signed URL
func (s *service) downloadVersion(ctx context.Context, versionID string) ([]byte, error) {
	tarball, err := s.db.getTarball(ctx, versionID)
	if err != nil {
		s.Error(err, "downloading module", "module_version_id", versionID)
		return nil, err
	}
	s.V(2).Info("downloaded module", "module_version_id", versionID)
	return tarball, nil
}

//lint:ignore U1000 to be used later
func (s *service) deleteVersion(ctx context.Context, versionID string) (*Module, error) {
	module, err := s.db.getModuleByID(ctx, versionID)
	if err != nil {
		s.Error(err, "retrieving module", "id", versionID)
		return nil, err
	}

	subject, err := s.organization.CanAccess(ctx, rbac.DeleteModuleVersionAction, module.Organization)
	if err != nil {
		return nil, err
	}

	if err = s.db.deleteModuleVersion(ctx, versionID); err != nil {
		s.Error(err, "deleting module", "subject", subject, "module", module)
		return nil, err
	}
	s.V(2).Info("deleted module", "subject", subject, "module", module)

	// return module w/o deleted version
	return s.db.getModuleByID(ctx, module.ID)
}
