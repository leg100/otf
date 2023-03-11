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
	"github.com/leg100/otf/rbac"
	"github.com/leg100/otf/semver"
	"github.com/leg100/surl"
)

type (
	// service is the service for modules
	service interface {
		// PublishModule publishes a module from a VCS repository.
		PublishModule(context.Context, PublishModuleOptions) (*Module, error)
		PublishVersion(context.Context, PublishModuleVersionOptions) error
		// CreateModule creates a module without a connection to a VCS repository.
		CreateModule(context.Context, CreateModuleOptions) (*Module, error)
		UpdateModuleStatus(ctx context.Context, opts UpdateModuleStatusOptions) (*Module, error)
		ListModules(context.Context, ListModulesOptions) ([]*Module, error)
		GetModule(ctx context.Context, opts GetModuleOptions) (*Module, error)
		GetModuleByID(ctx context.Context, id string) (*Module, error)
		GetModuleByRepoID(ctx context.Context, repoID uuid.UUID) (*Module, error)
		DeleteModule(ctx context.Context, id string) (*Module, error)
		GetModuleInfo(ctx context.Context, versionID string) (*TerraformModule, error)

		CreateVersion(context.Context, CreateModuleVersionOptions) (*ModuleVersion, error)

		uploadVersion(ctx context.Context, versionID string, tarball []byte) error
		downloadVersion(ctx context.Context, versionID string) ([]byte, error)
	}

	Service struct {
		otf.VCSProviderService
		logr.Logger
		*Publisher

		db       *pgdb
		repo     otf.RepoService
		hostname string

		organization otf.Authorizer

		api *api
		web *webHandlers
	}

	Options struct {
		OrganizationAuthorizer otf.Authorizer
		CloudService           cloud.Service
		Hostname               string

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
	svc.web = &webHandlers{
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

// PublishModule publishes a new module from a VCS repository, enumerating through
// its git tags and releasing a module version for each tag.
func (s *Service) PublishModule(ctx context.Context, opts PublishModuleOptions) (*Module, error) {
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

func (s *Service) publishModule(ctx context.Context, organization string, opts PublishModuleOptions) (*Module, error) {

	name, provider, err := opts.Repo.Split()
	if err != nil {
		return nil, err
	}

	mod := NewModule(CreateModuleOptions{
		Name:         name,
		Provider:     provider,
		Organization: organization,
	})

	// persist module to db and connect to repository
	err = s.db.tx(ctx, func(tx *pgdb) error {
		if err := tx.CreateModule(ctx, mod); err != nil {
			return err
		}
		connection, err := s.repo.Connect(ctx, otf.ConnectOptions{
			ConnectionType: otf.ModuleConnection,
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
		return s.UpdateModuleStatus(ctx, UpdateModuleStatusOptions{
			ID:     mod.ID,
			Status: ModuleStatusNoVersionTags,
		})
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
		err := s.PublishVersion(ctx, PublishModuleVersionOptions{
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
func (s *Service) PublishVersion(ctx context.Context, opts PublishModuleVersionOptions) error {
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
		return s.db.UpdateModuleVersionStatus(ctx, UpdateModuleVersionStatusOptions{
			ID:     modver.ID,
			Status: ModuleVersionStatusCloneFailed,
			Error:  err.Error(),
		})
	}

	return s.uploadVersion(ctx, modver.ID, tarball)
}

func (a *Service) CreateModule(ctx context.Context, opts CreateModuleOptions) (*Module, error) {
	subject, err := a.organization.CanAccess(ctx, rbac.CreateModuleAction, opts.Organization)
	if err != nil {
		return nil, err
	}

	module := NewModule(opts)

	if err := a.db.CreateModule(ctx, module); err != nil {
		a.Error(err, "creating module", "subject", subject, "module", module)
		return nil, err
	}
	a.V(0).Info("created module", "subject", subject, "module", module)
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

	subject, err := a.organization.CanAccess(ctx, rbac.GetModuleAction, module.Organization)
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
		return tx.delete(ctx, id)
	})
	if err != nil {
		a.Error(err, "deleting module", "subject", subject, "module", module)
		return nil, err
	}
	a.V(2).Info("deleted module", "subject", subject, "module", module)
	return module, nil
}

func (a *Service) CreateVersion(ctx context.Context, opts CreateModuleVersionOptions) (*ModuleVersion, error) {
	module, err := a.db.GetModuleByID(ctx, opts.ModuleID)
	if err != nil {
		return nil, err
	}

	subject, err := a.organization.CanAccess(ctx, rbac.CreateModuleVersionAction, module.Organization)
	if err != nil {
		return nil, err
	}

	modver := NewModuleVersion(opts)

	if err := a.db.CreateModuleVersion(ctx, modver); err != nil {
		a.Error(err, "creating module version", "organization", module.Organization, "subject", subject, "module_version", modver)
		return nil, err
	}
	a.V(0).Info("created module version", "organization", module.Organization, "subject", subject, "module_version", modver)
	return modver, nil
}

func (s *Service) GetModuleInfo(ctx context.Context, versionID string) (*TerraformModule, error) {
	tarball, err := s.db.getTarball(ctx, versionID)
	if err != nil {
		return nil, err
	}
	return unmarshalTerraformModule(tarball)
}

func (a *Service) uploadVersion(ctx context.Context, versionID string, tarball []byte) error {
	module, err := a.db.getModuleByVersionID(ctx, versionID)
	if err != nil {
		return err
	}

	// validate tarball
	if _, err := unmarshalTerraformModule(tarball); err != nil {
		a.Error(err, "uploading module version", "module_version", versionID)
		return a.db.UpdateModuleVersionStatus(ctx, UpdateModuleVersionStatusOptions{
			ID:     versionID,
			Status: ModuleVersionStatusRegIngressFailed,
			Error:  err.Error(),
		})
	}

	// save tarball, set status, and make it the latest version
	err = a.db.tx(ctx, func(tx *pgdb) error {
		if err := tx.saveTarball(ctx, versionID, tarball); err != nil {
			return err
		}
		err = tx.UpdateModuleVersionStatus(ctx, UpdateModuleVersionStatusOptions{
			ID:     versionID,
			Status: ModuleVersionStatusOK,
		})
		if err != nil {
			return err
		}
		// re-retrieve module so that includes the above version with updated
		// status
		module, err = tx.GetModuleByID(ctx, module.ID)
		if err != nil {
			return err
		}
		return nil
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

func (a *Service) deleteVersion(ctx context.Context, versionID string) (*Module, error) {
	module, err := a.db.GetModuleByID(ctx, versionID)
	if err != nil {
		a.Error(err, "retrieving module", "id", versionID)
		return nil, err
	}

	subject, err := a.organization.CanAccess(ctx, rbac.DeleteModuleVersionAction, module.Organization)
	if err != nil {
		return nil, err
	}

	if err = a.db.deleteModuleVersion(ctx, versionID); err != nil {
		a.Error(err, "deleting module", "subject", subject, "module", module)
		return nil, err
	}
	a.V(2).Info("deleted module", "subject", subject, "module", module)

	// return module w/o deleted version
	return a.db.GetModuleByID(ctx, module.ID)
}

// tx returns a service in a callback, with its database calls wrapped within a transaction
func (a *Service) tx(ctx context.Context, txFunc func(*Service) error) error {
	return a.db.Tx(ctx, func(db otf.DB) error {
		return txFunc(serviceWithDB(a, &pgdb{db}))
	})
}
