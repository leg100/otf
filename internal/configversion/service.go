package configversion

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/rbac"
	"github.com/leg100/surl"
)

type (
	// namespaced for disambiguation from other services
	ConfigurationVersionService = Service

	Service interface {
		CreateConfigurationVersion(ctx context.Context, workspaceID string, opts ConfigurationVersionCreateOptions) (*ConfigurationVersion, error)
		// CloneConfigurationVersion creates a new configuration version using the
		// config tarball of an existing configuration version.
		CloneConfigurationVersion(ctx context.Context, cvID string, opts ConfigurationVersionCreateOptions) (*ConfigurationVersion, error)
		GetConfigurationVersion(ctx context.Context, id string) (*ConfigurationVersion, error)
		GetLatestConfigurationVersion(ctx context.Context, workspaceID string) (*ConfigurationVersion, error)
		ListConfigurationVersions(ctx context.Context, workspaceID string, opts ConfigurationVersionListOptions) (*ConfigurationVersionList, error)

		// Upload handles verification and upload of the config tarball, updating
		// the config version upon success or failure.
		UploadConfig(ctx context.Context, id string, config []byte) error

		// Download retrieves the config tarball for the given config version ID.
		DownloadConfig(ctx context.Context, id string) ([]byte, error)

		DeleteConfigurationVersion(ctx context.Context, cvID string) error
	}

	service struct {
		logr.Logger

		workspace internal.Authorizer

		db    *pgdb
		cache internal.Cache
	}

	Options struct {
		logr.Logger

		WorkspaceAuthorizer internal.Authorizer

		internal.Cache
		internal.DB
		*surl.Signer
	}
)

func NewService(opts Options) *service {
	svc := service{
		Logger: opts.Logger,
	}

	svc.workspace = opts.WorkspaceAuthorizer

	svc.db = &pgdb{opts.DB}
	svc.cache = opts.Cache

	return &svc
}

func (s *service) CreateConfigurationVersion(ctx context.Context, workspaceID string, opts ConfigurationVersionCreateOptions) (*ConfigurationVersion, error) {
	subject, err := s.workspace.CanAccess(ctx, rbac.CreateConfigurationVersionAction, workspaceID)
	if err != nil {
		return nil, err
	}

	cv, err := NewConfigurationVersion(workspaceID, opts)
	if err != nil {
		s.Error(err, "constructing configuration version", "id", cv.ID, "subject", subject)
		return nil, err
	}
	if err := s.db.CreateConfigurationVersion(ctx, cv); err != nil {
		s.Error(err, "creating configuration version", "id", cv.ID, "subject", subject)
		return nil, err
	}
	s.V(2).Info("created configuration version", "id", cv.ID, "subject", subject)
	return cv, nil
}

func (s *service) CloneConfigurationVersion(ctx context.Context, cvID string, opts ConfigurationVersionCreateOptions) (*ConfigurationVersion, error) {
	cv, err := s.GetConfigurationVersion(ctx, cvID)
	if err != nil {
		return nil, err
	}

	cv, err = s.CreateConfigurationVersion(ctx, cv.WorkspaceID, opts)
	if err != nil {
		return nil, err
	}

	config, err := s.DownloadConfig(ctx, cvID)
	if err != nil {
		return nil, err
	}

	if err := s.UploadConfig(ctx, cv.ID, config); err != nil {
		return nil, err
	}

	return cv, nil
}

func (s *service) ListConfigurationVersions(ctx context.Context, workspaceID string, opts ConfigurationVersionListOptions) (*ConfigurationVersionList, error) {
	subject, err := s.workspace.CanAccess(ctx, rbac.ListConfigurationVersionsAction, workspaceID)
	if err != nil {
		return nil, err
	}

	cvl, err := s.db.ListConfigurationVersions(ctx, workspaceID, ConfigurationVersionListOptions{ListOptions: opts.ListOptions})
	if err != nil {
		s.Error(err, "listing configuration versions")
		return nil, err
	}

	s.V(2).Info("listed configuration versions", "subject", subject)
	return cvl, nil
}

func (s *service) GetConfigurationVersion(ctx context.Context, cvID string) (*ConfigurationVersion, error) {
	subject, err := s.canAccess(ctx, rbac.GetConfigurationVersionAction, cvID)
	if err != nil {
		return nil, err
	}

	cv, err := s.db.GetConfigurationVersion(ctx, ConfigurationVersionGetOptions{ID: &cvID})
	if err != nil {
		s.Error(err, "retrieving configuration version", "id", cvID, "subject", subject)
		return nil, err
	}
	s.V(2).Info("retrieved configuration version", "id", cvID, "subject", subject)
	return cv, nil
}

func (s *service) GetLatestConfigurationVersion(ctx context.Context, workspaceID string) (*ConfigurationVersion, error) {
	subject, err := s.workspace.CanAccess(ctx, rbac.GetConfigurationVersionAction, workspaceID)
	if err != nil {
		return nil, err
	}

	cv, err := s.db.GetConfigurationVersion(ctx, ConfigurationVersionGetOptions{WorkspaceID: &workspaceID})
	if err != nil {
		s.Error(err, "retrieving latest configuration version", "workspace_id", workspaceID, "subject", subject)
		return nil, err
	}
	s.V(2).Info("retrieved latest configuration version", "workspace_id", workspaceID, "subject", subject)
	return cv, nil
}

func (s *service) DeleteConfigurationVersion(ctx context.Context, cvID string) error {
	subject, err := s.canAccess(ctx, rbac.DeleteConfigurationVersionAction, cvID)
	if err != nil {
		return err
	}

	err = s.db.DeleteConfigurationVersion(ctx, cvID)
	if err != nil {
		s.Error(err, "deleting configuration version", "id", cvID, "subject", subject)
		return err
	}
	s.V(2).Info("deleted configuration version", "id", cvID, "subject", subject)
	return nil
}

func (s *service) canAccess(ctx context.Context, action rbac.Action, cvID string) (internal.Subject, error) {
	cv, err := s.db.GetConfigurationVersion(ctx, ConfigurationVersionGetOptions{ID: &cvID})
	if err != nil {
		return nil, err
	}
	return s.workspace.CanAccess(ctx, action, cv.WorkspaceID)
}
