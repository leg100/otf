package app

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
	"github.com/leg100/otf/sql"
)

var _ otf.ConfigurationVersionService = (*ConfigurationVersionService)(nil)

type ConfigurationVersionService struct {
	db    *sql.DB
	cache otf.Cache
	logr.Logger
}

func NewConfigurationVersionService(db *sql.DB, logger logr.Logger, cache otf.Cache) *ConfigurationVersionService {
	return &ConfigurationVersionService{
		db:     db,
		cache:  cache,
		Logger: logger,
	}
}

func (s ConfigurationVersionService) Create(ctx context.Context, workspaceID string, opts otf.ConfigurationVersionCreateOptions) (*otf.ConfigurationVersion, error) {
	cv, err := otf.NewConfigurationVersion(workspaceID, opts)
	if err != nil {
		s.Error(err, "constructing configuration version", "id", cv.ID())
		return nil, err
	}
	if err := s.db.CreateConfigurationVersion(ctx, cv); err != nil {
		s.Error(err, "creating configuration version", "id", cv.ID())
		return nil, err
	}
	s.V(2).Info("created configuration version", "id", cv.ID())
	return cv, nil
}

func (s ConfigurationVersionService) List(ctx context.Context, workspaceID string, opts otf.ConfigurationVersionListOptions) (*otf.ConfigurationVersionList, error) {
	cvl, err := s.db.ListConfigurationVersions(ctx, workspaceID, otf.ConfigurationVersionListOptions{ListOptions: opts.ListOptions})
	if err != nil {
		s.Error(err, "listing configuration versions")
		return nil, err
	}
	s.V(2).Info("listed configuration versions")
	return cvl, nil
}

func (s ConfigurationVersionService) Get(ctx context.Context, cvID string) (*otf.ConfigurationVersion, error) {
	cv, err := s.db.GetConfigurationVersion(ctx, otf.ConfigurationVersionGetOptions{ID: &cvID})
	if err != nil {
		s.Error(err, "retrieving configuration version", "id", cvID)
		return nil, err
	}
	s.V(2).Info("retrieved configuration version", "id", cvID)
	return cv, nil
}

func (s ConfigurationVersionService) GetLatest(ctx context.Context, workspaceID string) (*otf.ConfigurationVersion, error) {
	cv, err := s.db.GetConfigurationVersion(ctx, otf.ConfigurationVersionGetOptions{WorkspaceID: &workspaceID})
	if err != nil {
		s.Error(err, "retrieving latest configuration version", "workspace_id", workspaceID)
		return nil, err
	}
	s.V(2).Info("retrieved latest configuration version", "workspace_id", workspaceID)
	return cv, nil
}

// Upload a configuration version tarball
func (s ConfigurationVersionService) Upload(ctx context.Context, cvID string, config []byte) error {
	err := s.db.UploadConfigurationVersion(context.Background(), cvID, func(cv *otf.ConfigurationVersion, uploader otf.ConfigUploader) error {
		return cv.Upload(context.Background(), config, uploader)
	})
	if err != nil {
		s.Error(err, "uploading configuration")
		return err
	}
	if err := s.cache.Set(otf.ConfigVersionCacheKey(cvID), config); err != nil {
		return fmt.Errorf("caching configuration version tarball: %w", err)
	}
	if err != nil {
		s.Error(err, "uploading configuration")
		return err
	}
	s.V(2).Info("uploaded configuration", "id", cvID, "bytes", len(config))
	return nil
}

func (s ConfigurationVersionService) Download(ctx context.Context, cvID string) ([]byte, error) {
	if config, err := s.cache.Get(otf.ConfigVersionCacheKey(cvID)); err == nil {
		return config, nil
	}
	config, err := s.db.GetConfig(context.Background(), cvID)
	if err != nil {
		return nil, err
	}
	if err := s.cache.Set(otf.ConfigVersionCacheKey(cvID), config); err != nil {
		return nil, fmt.Errorf("caching configuration version tarball: %w", err)
	}
	s.V(2).Info("uploaded configuration", "id", cvID, "bytes", len(config))
	return config, nil
}
