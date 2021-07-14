package mock

import (
	"github.com/leg100/go-tfe"
	"github.com/leg100/ots"
)

var _ ots.StateVersionService = (*StateVersionService)(nil)

type StateVersionService struct {
	CreateFn   func(workspaceID string, opts tfe.StateVersionCreateOptions) (*ots.StateVersion, error)
	CurrentFn  func(workspaceID string) (*ots.StateVersion, error)
	GetFn      func(id string) (*ots.StateVersion, error)
	DownloadFn func(id string) ([]byte, error)
	ListFn     func(opts tfe.StateVersionListOptions) (*ots.StateVersionList, error)
}

func (s StateVersionService) Create(workspaceID string, opts tfe.StateVersionCreateOptions) (*ots.StateVersion, error) {
	return s.CreateFn(workspaceID, opts)
}

func (s StateVersionService) Get(id string) (*ots.StateVersion, error) {
	return s.GetFn(id)
}

func (s StateVersionService) Current(workspaceID string) (*ots.StateVersion, error) {
	return s.CurrentFn(workspaceID)
}

func (s StateVersionService) Download(id string) ([]byte, error) {
	return s.DownloadFn(id)
}

func (s StateVersionService) List(opts tfe.StateVersionListOptions) (*ots.StateVersionList, error) {
	return s.ListFn(opts)
}
