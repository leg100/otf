package app

import (
	"github.com/leg100/go-tfe"
	"github.com/leg100/otf"
)

var _ otf.StateVersionService = (*StateVersionService)(nil)

type StateVersionService struct {
	db otf.StateVersionStore
	bs otf.BlobStore
	*otf.StateVersionFactory
}

func NewStateVersionService(db otf.StateVersionStore, wss otf.WorkspaceService, bs otf.BlobStore) *StateVersionService {
	return &StateVersionService{
		bs: bs,
		db: db,
		StateVersionFactory: &otf.StateVersionFactory{
			BlobStore:        bs,
			WorkspaceService: wss,
		},
	}
}

func (s StateVersionService) Create(workspaceID string, opts tfe.StateVersionCreateOptions) (*otf.StateVersion, error) {
	sv, err := s.NewStateVersion(workspaceID, opts)
	if err != nil {
		return nil, err
	}

	return s.db.Create(sv)
}

func (s StateVersionService) List(opts tfe.StateVersionListOptions) (*otf.StateVersionList, error) {
	return s.db.List(opts)
}

func (s StateVersionService) Current(workspaceID string) (*otf.StateVersion, error) {
	return s.db.Get(otf.StateVersionGetOptions{WorkspaceID: &workspaceID})
}

func (s StateVersionService) Get(id string) (*otf.StateVersion, error) {
	return s.db.Get(otf.StateVersionGetOptions{ID: &id})
}

func (s StateVersionService) Download(id string) ([]byte, error) {
	sv, err := s.db.Get(otf.StateVersionGetOptions{ID: &id})
	if err != nil {
		return nil, err
	}

	return s.bs.Get(sv.BlobID)
}
