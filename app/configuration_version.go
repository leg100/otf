package app

import (
	"github.com/leg100/go-tfe"
	"github.com/leg100/ots"
)

var _ ots.ConfigurationVersionService = (*ConfigurationVersionService)(nil)

type ConfigurationVersionService struct {
	db ots.ConfigurationVersionStore
	bs ots.BlobStore

	*ots.ConfigurationVersionFactory
}

func NewConfigurationVersionService(db ots.ConfigurationVersionStore, wss ots.WorkspaceService, bs ots.BlobStore) *ConfigurationVersionService {
	return &ConfigurationVersionService{
		bs: bs,
		db: db,
		ConfigurationVersionFactory: &ots.ConfigurationVersionFactory{
			WorkspaceService: wss,
		},
	}
}

func (s ConfigurationVersionService) Create(workspaceID string, opts *tfe.ConfigurationVersionCreateOptions) (*ots.ConfigurationVersion, error) {
	cv, err := s.NewConfigurationVersion(workspaceID, opts)
	if err != nil {
		return nil, err
	}

	return s.db.Create(cv)
}

func (s ConfigurationVersionService) List(workspaceID string, opts tfe.ConfigurationVersionListOptions) (*ots.ConfigurationVersionList, error) {
	return s.db.List(workspaceID, ots.ConfigurationVersionListOptions{ListOptions: opts.ListOptions})
}

func (s ConfigurationVersionService) Get(id string) (*ots.ConfigurationVersion, error) {
	return s.db.Get(ots.ConfigurationVersionGetOptions{ID: &id})
}

func (s ConfigurationVersionService) GetLatest(workspaceID string) (*ots.ConfigurationVersion, error) {
	return s.db.Get(ots.ConfigurationVersionGetOptions{WorkspaceID: &workspaceID})
}

// Upload a configuration version blob
func (s ConfigurationVersionService) Upload(id string, configuration []byte) error {
	blobID, err := s.bs.Put(configuration)
	if err != nil {
		return err
	}

	_, err = s.db.Update(id, func(cv *ots.ConfigurationVersion) error {
		cv.BlobID = blobID
		cv.Status = tfe.ConfigurationUploaded

		return nil
	})

	return err
}

func (s ConfigurationVersionService) Download(id string) ([]byte, error) {
	cv, err := s.db.Get(ots.ConfigurationVersionGetOptions{ID: &id})
	if err != nil {
		return nil, err
	}

	return s.bs.Get(cv.BlobID)
}
