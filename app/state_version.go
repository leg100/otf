package app

import (
	"github.com/go-logr/logr"
	"github.com/leg100/otf"
)

var _ otf.StateVersionService = (*StateVersionService)(nil)

type StateVersionService struct {
	db otf.StateVersionStore
	bs otf.BlobStore
	*otf.StateVersionFactory

	logr.Logger
}

func NewStateVersionService(db otf.StateVersionStore, logger logr.Logger, wss otf.WorkspaceService, bs otf.BlobStore) *StateVersionService {
	return &StateVersionService{
		bs:     bs,
		db:     db,
		Logger: logger,
		StateVersionFactory: &otf.StateVersionFactory{
			BlobStore:        bs,
			WorkspaceService: wss,
		},
	}
}

func (s StateVersionService) Create(workspaceID string, opts otf.StateVersionCreateOptions) (*otf.StateVersion, error) {
	sv, err := s.NewStateVersion(workspaceID, opts)
	if err != nil {
		return nil, err
	}

	_, err = s.db.Create(sv)
	if err != nil {
		s.Error(err, "creating state version")
		return nil, err
	}

	s.V(0).Info("created state version", "id", sv.ID, "workspace", sv.Workspace.Name, "serial", sv.Serial)

	return sv, nil
}

func (s StateVersionService) List(opts otf.StateVersionListOptions) (*otf.StateVersionList, error) {
	return s.db.List(opts)
}

func (s StateVersionService) Current(workspaceID string) (*otf.StateVersion, error) {
	sv, err := s.db.Get(otf.StateVersionGetOptions{WorkspaceID: &workspaceID})
	if err != nil {
		s.Error(err, "retrieving current state version", "workspace_id", workspaceID)
		return nil, err
	}

	return sv, nil
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
