package mock

import (
	"github.com/leg100/go-tfe"
	"github.com/leg100/otf"
)

var _ otf.ConfigurationVersionService = (*ConfigurationVersionService)(nil)

type ConfigurationVersionService struct {
	CreateFn    func(workspaceID string, opts *tfe.ConfigurationVersionCreateOptions) (*otf.ConfigurationVersion, error)
	GetFn       func(id string) (*otf.ConfigurationVersion, error)
	GetLatestFn func(workspaceID string) (*otf.ConfigurationVersion, error)
	ListFn      func(workspaceID string, opts tfe.ConfigurationVersionListOptions) (*otf.ConfigurationVersionList, error)
	UploadFn    func(id string, payload []byte) error
	DownloadFn  func(id string) ([]byte, error)
}

func (s ConfigurationVersionService) Create(workspaceID string, opts *tfe.ConfigurationVersionCreateOptions) (*otf.ConfigurationVersion, error) {
	return s.CreateFn(workspaceID, opts)
}

func (s ConfigurationVersionService) Get(id string) (*otf.ConfigurationVersion, error) {
	return s.GetFn(id)
}

func (s ConfigurationVersionService) GetLatest(workspaceID string) (*otf.ConfigurationVersion, error) {
	return s.GetLatestFn(workspaceID)
}

func (s ConfigurationVersionService) List(workspaceID string, opts tfe.ConfigurationVersionListOptions) (*otf.ConfigurationVersionList, error) {
	return s.ListFn(workspaceID, opts)
}

func (s ConfigurationVersionService) Upload(id string, payload []byte) error {
	return s.UploadFn(id, payload)
}

func (s ConfigurationVersionService) Download(id string) ([]byte, error) {
	return s.DownloadFn(id)
}
