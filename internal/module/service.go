package module

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/leg100/otf/internal/logr"
	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/connections"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/repohooks"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/semver"
	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/vcs"
	"github.com/leg100/surl/v2"
)

type (
	Service struct {
		logr.Logger
		*publisher
		*authz.Authorizer

		db *pgdb

		api          *api
		vcsproviders *vcs.Service
		connections  *connections.Service
	}

	Options struct {
		logr.Logger

		*sql.DB
		*internal.HostnameService
		*surl.Signer

		Authorizer         *authz.Authorizer
		RepohookService    *repohooks.Service
		VCSProviderService *vcs.Service
		ConnectionsService *connections.Service
		VCSEventSubscriber vcs.Subscriber
	}
)

func NewService(opts Options) *Service {
	svc := Service{
		Logger:       opts.Logger,
		Authorizer:   opts.Authorizer,
		connections:  opts.ConnectionsService,
		db:           &pgdb{opts.DB},
		vcsproviders: opts.VCSProviderService,
	}
	svc.api = &api{
		svc:    &svc,
		Signer: opts.Signer,
	}
	publisher := &publisher{
		Logger:       opts.Logger.WithValues("component", "publisher"),
		vcsproviders: opts.VCSProviderService,
		modules:      &svc,
	}
	// Subscribe module publisher to incoming vcs events
	opts.VCSEventSubscriber.Subscribe(publisher.handle)

	return &svc
}

func (s *Service) AddHandlers(r *mux.Router) {
	s.api.addHandlers(r)
}

// PublishModule publishes a new module from a VCS repository, enumerating through
// its git tags and releasing a module version for each tag.
func (s *Service) PublishModule(ctx context.Context, opts PublishOptions) (*Module, error) {
	vcsprov, err := s.vcsproviders.Get(ctx, opts.VCSProviderID)
	if err != nil {
		return nil, err
	}

	subject, err := s.Authorize(ctx, authz.CreateModuleAction, &vcsprov.Organization)
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

func (s *Service) publishModule(ctx context.Context, organization organization.Name, opts PublishOptions) (*Module, error) {
	name, provider, err := opts.Repo.Split()
	if err != nil {
		return nil, err
	}

	mod := newModule(CreateOptions{
		Name:         name,
		Provider:     provider,
		Organization: organization,
	})

	// persist module to db and connect to repository
	if err := s.db.createModule(ctx, mod); err != nil {
		return nil, fmt.Errorf("creating module in database: %w", err)
	}

	var (
		client vcs.Client
		tags   []string
	)
	err = func() (err error) {
		mod.Connection, err = s.connections.Connect(ctx, connections.ConnectOptions{
			ResourceID:    mod.ID,
			VCSProviderID: opts.VCSProviderID,
			RepoPath:      vcs.Repo(opts.Repo),
		})
		if err != nil {
			return err
		}
		client, err = s.vcsproviders.Get(ctx, opts.VCSProviderID)
		if err != nil {
			return fmt.Errorf("retreving vcs client config: %w", err)
		}
		tags, err = client.ListTags(ctx, vcs.ListTagsOptions{
			Repo: vcs.Repo(opts.Repo),
		})
		if err != nil {
			return err
		}
		return nil
	}()
	if err != nil {
		mod, updateErr := s.updateModuleStatus(ctx, mod, ModuleStatusSetupFailed)
		return mod, errors.Join(err, updateErr)
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
	if _, err := s.updateModuleStatus(ctx, mod, ModuleStatusSetupComplete); err != nil {
		return nil, err
	}
	// return module from db complete with versions
	return s.GetModuleByID(ctx, mod.ID)
}

// PublishVersion publishes a module version, retrieving its contents from a repository and
// uploading it to the module store.
func (s *Service) PublishVersion(ctx context.Context, opts PublishVersionOptions) error {
	modver, err := s.CreateVersion(ctx, CreateModuleVersionOptions{
		ModuleID: opts.ModuleID,
		Version:  opts.Version,
	})
	if err != nil {
		return err
	}

	tarball, _, err := opts.Client.GetRepoTarball(ctx, vcs.GetRepoTarballOptions{
		Repo: vcs.Repo(opts.Repo),
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

func (s *Service) CreateModule(ctx context.Context, opts CreateOptions) (*Module, error) {
	subject, err := s.Authorize(ctx, authz.CreateModuleAction, &opts.Organization)
	if err != nil {
		return nil, err
	}

	module := newModule(opts)

	if err := s.db.createModule(ctx, module); err != nil {
		s.Error(err, "creating module", "subject", subject, "module", module)
		return nil, err
	}
	s.V(0).Info("created module", "subject", subject, "module", module)
	return module, nil
}

func (s *Service) ListModules(ctx context.Context, opts ListOptions) ([]*Module, error) {
	subject, err := s.Authorize(ctx, authz.ListModulesAction, &opts.Organization)
	if err != nil {
		return nil, err
	}

	page, err := s.db.listModules(ctx, opts)
	if err != nil {
		s.Error(err, "listing modules", "organization", opts.Organization, "subject", subject)
		return nil, err
	}
	s.V(9).Info("listed modules", "organization", opts.Organization, "subject", subject)
	return page, nil
}

func (s *Service) ListProviders(ctx context.Context, organization organization.Name) ([]string, error) {
	return s.db.listProviders(ctx, organization)
}

func (s *Service) GetModule(ctx context.Context, opts GetModuleOptions) (*Module, error) {
	subject, err := s.Authorize(ctx, authz.GetModuleAction, &opts.Organization)
	if err != nil {
		return nil, err
	}

	module, err := s.db.getModule(ctx, opts)
	if err != nil {
		s.Error(err, "retrieving module", "module", opts)
		return nil, err
	}

	s.V(9).Info("retrieved module", "subject", subject, "module", module)
	return module, nil
}

func (s *Service) GetModuleByID(ctx context.Context, id resource.TfeID) (*Module, error) {
	module, err := s.db.getModuleByID(ctx, id)
	if err != nil {
		s.Error(err, "retrieving module", "id", id)
		return nil, err
	}

	subject, err := s.Authorize(ctx, authz.GetModuleAction, &module.Organization)
	if err != nil {
		return nil, err
	}

	s.V(9).Info("retrieved module", "subject", subject, "module", module)
	return module, nil
}

func (s *Service) GetModuleByConnection(ctx context.Context, vcsProviderID resource.TfeID, repoPath vcs.Repo) (*Module, error) {
	return s.db.getModuleByConnection(ctx, vcsProviderID, repoPath)
}

func (s *Service) DeleteModule(ctx context.Context, id resource.TfeID) (*Module, error) {
	module, err := s.db.getModuleByID(ctx, id)
	if err != nil {
		s.Error(err, "retrieving module", "id", id)
		return nil, err
	}

	subject, err := s.Authorize(ctx, authz.DeleteModuleAction, &module.Organization)
	if err != nil {
		return nil, err
	}

	err = s.db.Tx(ctx, func(ctx context.Context) error {
		// disconnect module prior to deletion
		if module.Connection != nil {
			err := s.connections.Disconnect(ctx, connections.DisconnectOptions{
				ResourceID: module.ID,
			})
			if err != nil {
				return err
			}
		}
		return s.db.delete(ctx, id)
	})
	if err != nil {
		s.Error(err, "deleting module", "subject", subject, "module", module)
		return nil, err
	}
	s.V(2).Info("deleted module", "subject", subject, "module", module)
	return module, nil
}

func (s *Service) CreateVersion(ctx context.Context, opts CreateModuleVersionOptions) (*ModuleVersion, error) {
	module, err := s.db.getModuleByID(ctx, opts.ModuleID)
	if err != nil {
		return nil, err
	}

	subject, err := s.Authorize(ctx, authz.CreateModuleVersionAction, &module.Organization)
	if err != nil {
		return nil, err
	}

	modver := newModuleVersion(opts)

	if err := s.db.createModuleVersion(ctx, modver); err != nil {
		s.Error(err, "creating module version", "organization", module.Organization, "subject", subject, "module_version", modver)
		return nil, err
	}
	s.V(0).Info("created module version", "organization", module.Organization, "subject", subject, "module_version", modver)
	return modver, nil
}

func (s *Service) GetModuleInfo(ctx context.Context, versionID resource.TfeID) (*TerraformModule, error) {
	tarball, err := s.db.getTarball(ctx, versionID)
	if err != nil {
		return nil, err
	}
	return UnmarshalTerraformModule(tarball)
}

func (s *Service) updateModuleStatus(ctx context.Context, mod *Module, status ModuleStatus) (*Module, error) {
	mod.Status = status

	if err := s.db.updateModuleStatus(ctx, mod.ID, status); err != nil {
		s.Error(err, "updating module status", "module", mod.ID, "status", status)
		return nil, err
	}
	s.V(0).Info("updated module status", "module", mod.ID, "status", status)
	return mod, nil
}

func (s *Service) uploadVersion(ctx context.Context, versionID resource.TfeID, tarball []byte) error {
	module, err := s.db.getModuleByVersionID(ctx, versionID)
	if err != nil {
		return err
	}

	// validate tarball
	if _, err := UnmarshalTerraformModule(tarball); err != nil {
		s.Error(err, "uploading module version", "module_version", versionID)
		return s.db.updateModuleVersionStatus(ctx, UpdateModuleVersionStatusOptions{
			ID:     versionID,
			Status: ModuleVersionStatusRegIngressFailed,
			Error:  err.Error(),
		})
	}

	// save tarball, set status, and make it the latest version
	err = s.db.Tx(ctx, func(ctx context.Context) error {
		if err := s.db.saveTarball(ctx, versionID, tarball); err != nil {
			return err
		}
		err := s.db.updateModuleVersionStatus(ctx, UpdateModuleVersionStatusOptions{
			ID:     versionID,
			Status: ModuleVersionStatusOK,
		})
		if err != nil {
			return err
		}
		// re-retrieve module so that includes the above version with updated
		// status
		_, err = s.db.getModuleByID(ctx, module.ID)
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
func (s *Service) downloadVersion(ctx context.Context, versionID resource.TfeID) ([]byte, error) {
	tarball, err := s.db.getTarball(ctx, versionID)
	if err != nil {
		s.Error(err, "downloading module", "module_version_id", versionID)
		return nil, err
	}
	s.V(9).Info("downloaded module", "module_version_id", versionID)
	return tarball, nil
}

//lint:ignore U1000 to be used later
func (s *Service) deleteVersion(ctx context.Context, versionID resource.TfeID) (*Module, error) {
	module, err := s.db.getModuleByID(ctx, versionID)
	if err != nil {
		s.Error(err, "retrieving module", "id", versionID)
		return nil, err
	}

	subject, err := s.Authorize(ctx, authz.DeleteModuleVersionAction, &module.Organization)
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
