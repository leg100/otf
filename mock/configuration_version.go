package mock

import (
	"github.com/leg100/go-tfe"
	"github.com/leg100/ots"
)

var _ ots.ConfigurationVersionService = (*ConfigurationVersionService)(nil)

type ConfigurationVersionService struct {
	CreateFn    func(workspaceID string, opts *tfe.ConfigurationVersionCreateOptions) (*ots.ConfigurationVersion, error)
	GetFn       func(id string) (*ots.ConfigurationVersion, error)
	GetLatestFn func(workspaceID string) (*ots.ConfigurationVersion, error)
	ListFn      func(workspaceID string, opts tfe.ConfigurationVersionListOptions) (*ots.ConfigurationVersionList, error)
	UploadFn    func(id string, payload []byte) error
	DownloadFn  func(id string) ([]byte, error)
}

func (s ConfigurationVersionService) Create(workspaceID string, opts *tfe.ConfigurationVersionCreateOptions) (*ots.ConfigurationVersion, error) {
	return s.CreateFn(workspaceID, opts)
}

func (s ConfigurationVersionService) Get(id string) (*ots.ConfigurationVersion, error) {
	return s.GetFn(id)
}

func (s ConfigurationVersionService) GetLatest(workspaceID string) (*ots.ConfigurationVersion, error) {
	return s.GetLatestFn(workspaceID)
}

func (s ConfigurationVersionService) List(workspaceID string, opts tfe.ConfigurationVersionListOptions) (*ots.ConfigurationVersionList, error) {
	return s.ListFn(workspaceID, opts)
}

func (s ConfigurationVersionService) Upload(id string, payload []byte) error {
	return s.UploadFn(id, payload)
}

func (s ConfigurationVersionService) Download(id string) ([]byte, error) {
	return s.DownloadFn(id)
}
