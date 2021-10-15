package app

import (
	"github.com/go-logr/logr"
	"github.com/leg100/otf"
)

var _ otf.ConfigurationVersionService = (*ConfigurationVersionService)(nil)

type ConfigurationVersionService struct {
	db otf.ConfigurationVersionStore
	bs otf.BlobStore

	logr.Logger

	*otf.ConfigurationVersionFactory
}

func NewConfigurationVersionService(db otf.ConfigurationVersionStore, logger logr.Logger, wss otf.WorkspaceService, bs otf.BlobStore) *ConfigurationVersionService {
	return &ConfigurationVersionService{
		bs:     bs,
		db:     db,
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

	s.V(3).Info("created configuration version", "id", cv.ID)

	return cv, nil
}

func (s ConfigurationVersionService) List(workspaceID string, opts otf.ConfigurationVersionListOptions) (*otf.ConfigurationVersionList, error) {
	cvl, err := s.db.List(workspaceID, otf.ConfigurationVersionListOptions{ListOptions: opts.ListOptions})
	if err != nil {
		s.Error(err, "listing configuration versions")
		return nil, err
	}

	s.V(3).Info("listed configuration versions")

	return cvl, nil
}

func (s ConfigurationVersionService) Get(id string) (*otf.ConfigurationVersion, error) {
	cv, err := s.db.Get(otf.ConfigurationVersionGetOptions{ID: &id})
	if err != nil {
		s.Error(err, "retrieving configuration version", "id", id)
		return nil, err
	}

	s.V(3).Info("retrieved configuration version", "id", id)

	return cv, nil
}

func (s ConfigurationVersionService) GetLatest(workspaceID string) (*otf.ConfigurationVersion, error) {
	return s.db.Get(otf.ConfigurationVersionGetOptions{WorkspaceID: &workspaceID})
}

// Upload a configuration version blob
func (s ConfigurationVersionService) Upload(id string, configuration []byte) error {
	_, err := s.db.Update(id, func(cv *otf.ConfigurationVersion) error {
		if err := s.bs.Put(cv.BlobID, configuration); err != nil {
			return err
		}

		cv.Status = otf.ConfigurationUploaded

		return nil
	})

	return err
}

func (s ConfigurationVersionService) Download(id string) ([]byte, error) {
	cv, err := s.db.Get(otf.ConfigurationVersionGetOptions{ID: &id})
	if err != nil {
		return nil, err
	}

	return s.bs.Get(cv.BlobID)
}
