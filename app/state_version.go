package app

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
	"github.com/leg100/otf/sql"
)

var _ otf.StateVersionService = (*StateVersionService)(nil)

type StateVersionService struct {
	db    *sql.DB
	cache otf.Cache

	logr.Logger
}

func NewStateVersionService(db *sql.DB, logger logr.Logger, cache otf.Cache) *StateVersionService {
	return &StateVersionService{
		db:     db,
		cache:  cache,
		Logger: logger,
	}
}

func (s StateVersionService) Create(ctx context.Context, workspaceID string, opts otf.StateVersionCreateOptions) (*otf.StateVersion, error) {
	sv, err := otf.NewStateVersion(opts)
	if err != nil {
		s.Error(err, "constructing state version")
		return nil, err
	}
	if err := s.db.CreateStateVersion(ctx, workspaceID, sv); err != nil {
		s.Error(err, "creating state version")
		return nil, err
	}

	if err := s.cache.Set(otf.StateVersionCacheKey(sv.ID()), sv.State()); err != nil {
		return nil, fmt.Errorf("caching state version: %w", err)
	}

	s.V(0).Info("created state version", "id", sv.ID(), "workspace", workspaceID, "serial", sv.Serial())
	return sv, nil
}

func (s StateVersionService) List(ctx context.Context, opts otf.StateVersionListOptions) (*otf.StateVersionList, error) {
	svl, err := s.db.ListStateVersions(ctx, opts)
	if err != nil {
		s.Error(err, "listing state versions", opts.LogFields()...)
		return nil, err
	}
	s.V(2).Info("listed state version", opts.LogFields()...)
	return svl, nil
}

func (s StateVersionService) Current(ctx context.Context, workspaceID string) (*otf.StateVersion, error) {
	sv, err := s.db.GetStateVersion(ctx, otf.StateVersionGetOptions{WorkspaceID: &workspaceID})
	if err != nil {
		s.Error(err, "retrieving current state version", "workspace_id", workspaceID)
		return nil, err
	}
	s.V(2).Info("retrieved current state version", "workspace_id", workspaceID)
	return sv, nil
}

func (s StateVersionService) Get(ctx context.Context, id string) (*otf.StateVersion, error) {
	sv, err := s.db.GetStateVersion(ctx, otf.StateVersionGetOptions{ID: &id})
	if err != nil {
		s.Error(err, "retrieving state version", "id", id)
		return nil, err
	}
	s.V(2).Info("retrieved state version", "id", id)
	return sv, nil
}

// Download state itself.
func (s StateVersionService) Download(ctx context.Context, id string) ([]byte, error) {
	if state, err := s.cache.Get(otf.StateVersionCacheKey(id)); err == nil {
		s.V(2).Info("downloaded state", "id", id)
		return state, nil
	}
	state, err := s.db.GetState(ctx, id)
	if err != nil {
		s.Error(err, "downloading state", "id", id)
		return nil, err
	}
	if err := s.cache.Set(otf.StateVersionCacheKey(id), state); err != nil {
		return nil, fmt.Errorf("caching state: %w", err)
	}
	s.V(2).Info("downloaded state", "id", id)
	return state, nil
}
