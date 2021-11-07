package app

import (
	"fmt"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
)

var _ otf.StateVersionService = (*StateVersionService)(nil)

type StateVersionService struct {
	db otf.StateVersionStore
	*otf.StateVersionFactory

	cache otf.Cache

	logr.Logger
}

func NewStateVersionService(db otf.StateVersionStore, logger logr.Logger, wss otf.WorkspaceService, cache otf.Cache) *StateVersionService {
	return &StateVersionService{
		db:     db,
		cache:  cache,
		Logger: logger,
		StateVersionFactory: &otf.StateVersionFactory{
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

	if err := s.cache.Set(otf.StateVersionCacheKey(sv.ID), sv.State); err != nil {
		return nil, fmt.Errorf("caching state version: %w", err)
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
	if state, err := s.cache.Get(otf.StateVersionCacheKey(id)); err == nil {
		return state, nil
	}

	sv, err := s.db.Get(otf.StateVersionGetOptions{ID: &id, State: true})
	if err != nil {
		s.Error(err, "retrieving state version", "id", sv.ID, "workspace", sv.Workspace.Name, "serial", sv.Serial)
		return nil, err
	}

	if err := s.cache.Set(otf.StateVersionCacheKey(id), sv.State); err != nil {
		return nil, fmt.Errorf("caching state version: %w", err)
	}

	return sv.State, nil
}
