package mock

import (
	"github.com/leg100/go-tfe"
	"github.com/leg100/otf"
)

var _ otf.StateVersionService = (*StateVersionService)(nil)

type StateVersionService struct {
	CreateFn   func(workspaceID string, opts tfe.StateVersionCreateOptions) (*otf.StateVersion, error)
	CurrentFn  func(workspaceID string) (*otf.StateVersion, error)
	GetFn      func(id string) (*otf.StateVersion, error)
	DownloadFn func(id string) ([]byte, error)
	ListFn     func(opts tfe.StateVersionListOptions) (*otf.StateVersionList, error)
}

func (s StateVersionService) Create(workspaceID string, opts tfe.StateVersionCreateOptions) (*otf.StateVersion, error) {
	return s.CreateFn(workspaceID, opts)
}

func (s StateVersionService) Get(id string) (*otf.StateVersion, error) {
	return s.GetFn(id)
}

func (s StateVersionService) Current(workspaceID string) (*otf.StateVersion, error) {
	return s.CurrentFn(workspaceID)
}

func (s StateVersionService) Download(id string) ([]byte, error) {
	return s.DownloadFn(id)
}

func (s StateVersionService) List(opts tfe.StateVersionListOptions) (*otf.StateVersionList, error) {
	return s.ListFn(opts)
}
