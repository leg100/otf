package app

import (
	"fmt"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
)

var _ otf.ConfigurationVersionService = (*ConfigurationVersionService)(nil)

type ConfigurationVersionService struct {
	db otf.ConfigurationVersionStore

	cache otf.Cache

	logr.Logger

	*otf.ConfigurationVersionFactory
}

func NewConfigurationVersionService(db otf.ConfigurationVersionStore, logger logr.Logger, wss otf.WorkspaceService, cache otf.Cache) *ConfigurationVersionService {
	return &ConfigurationVersionService{
		db:     db,
		cache:  cache,
		Logger: logger,
		ConfigurationVersionFactory: &otf.ConfigurationVersionFactory{
			WorkspaceService: wss,
		},
	}
}

func (s ConfigurationVersionService) Create(workspaceID string, opts otf.ConfigurationVersionCreateOptions) (*otf.ConfigurationVersion, error) {
	cv, err := s.NewConfigurationVersion(workspaceID, opts)
	if err != nil {
		return nil, err
	}

	_, err = s.db.Create(cv)
	if err != nil {
		s.Error(err, "creating configuration version", "id", cv.ID)
		return nil, err
	}

	s.V(2).Info("created configuration version", "id", cv.ID)

	return cv, nil
}

func (s ConfigurationVersionService) List(workspaceID string, opts otf.ConfigurationVersionListOptions) (*otf.ConfigurationVersionList, error) {
	cvl, err := s.db.List(workspaceID, otf.ConfigurationVersionListOptions{ListOptions: opts.ListOptions})
	if err != nil {
		s.Error(err, "listing configuration versions")
		return nil, err
	}

	s.V(2).Info("listed configuration versions")

	return cvl, nil
}

func (s ConfigurationVersionService) Get(id string) (*otf.ConfigurationVersion, error) {
	cv, err := s.db.Get(otf.ConfigurationVersionGetOptions{ID: &id})
	if err != nil {
		s.Error(err, "retrieving configuration version", "id", id)
		return nil, err
	}

	s.V(2).Info("retrieved configuration version", "id", id)

	return cv, nil
}

func (s ConfigurationVersionService) GetLatest(workspaceID string) (*otf.ConfigurationVersion, error) {
	return s.db.Get(otf.ConfigurationVersionGetOptions{WorkspaceID: &workspaceID})
}

// Upload a configuration version tarball
func (s ConfigurationVersionService) Upload(id string, config []byte) error {
	_, err := s.db.Update(id, func(cv *otf.ConfigurationVersion) error {
		cv.Config = config
		cv.Status = otf.ConfigurationUploaded

		return nil
	})

	if err := s.cache.Set(otf.ConfigVersionCacheKey(id), config); err != nil {
		return fmt.Errorf("caching configuration version tarball: %w", err)
	}

	if err != nil {
		s.Error(err, "uploading configuration version")
		return err
	}

	return nil
}

func (s ConfigurationVersionService) Download(id string) ([]byte, error) {
	if config, err := s.cache.Get(otf.ConfigVersionCacheKey(id)); err == nil {
		return config, nil
	}

	cv, err := s.db.Get(otf.ConfigurationVersionGetOptions{ID: &id, Config: true})
	if err != nil {
		return nil, err
	}

	if err := s.cache.Set(otf.ConfigVersionCacheKey(id), cv.Config); err != nil {
		return nil, fmt.Errorf("caching configuration version tarball: %w", err)
	}

	return cv.Config, nil
}
