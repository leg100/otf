package app

import (
	"github.com/leg100/go-tfe"
	"github.com/leg100/ots"
)

var _ ots.StateVersionService = (*StateVersionService)(nil)

type StateVersionService struct {
	db ots.StateVersionStore
	bs ots.BlobStore
	*ots.StateVersionFactory
}

func NewStateVersionService(db ots.StateVersionStore, wss ots.WorkspaceService, bs ots.BlobStore) *StateVersionService {
	return &StateVersionService{
		bs: bs,
		db: db,
		StateVersionFactory: &ots.StateVersionFactory{
			BlobStore:        bs,
			WorkspaceService: wss,
		},
	}
}

func (s StateVersionService) Create(workspaceID string, opts tfe.StateVersionCreateOptions) (*ots.StateVersion, error) {
	sv, err := s.NewStateVersion(workspaceID, opts)
	if err != nil {
		return nil, err
	}

	return s.db.Create(sv)
}

func (s StateVersionService) List(opts tfe.StateVersionListOptions) (*ots.StateVersionList, error) {
	return s.db.List(opts)
}

func (s StateVersionService) Current(workspaceID string) (*ots.StateVersion, error) {
	return s.db.Get(ots.StateVersionGetOptions{WorkspaceID: &workspaceID})
}

func (s StateVersionService) Get(id string) (*ots.StateVersion, error) {
	return s.db.Get(ots.StateVersionGetOptions{ID: &id})
}

func (s StateVersionService) Download(id string) ([]byte, error) {
	sv, err := s.db.Get(ots.StateVersionGetOptions{ID: &id})
	if err != nil {
		return nil, err
	}

	return s.bs.Get(sv.BlobID)
}
