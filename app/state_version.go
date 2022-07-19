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

func (s StateVersionService) CreateStateVersion(ctx context.Context, workspaceID string, opts otf.StateVersionCreateOptions) (*otf.StateVersion, error) {
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

func (s StateVersionService) ListStateVersion(ctx context.Context, opts otf.StateVersionListOptions) (*otf.StateVersionList, error) {
	svl, err := s.db.ListStateVersions(ctx, opts)
	if err != nil {
		s.Error(err, "listing state versions", opts.LogFields()...)
		return nil, err
	}
	s.V(2).Info("listed state version", opts.LogFields()...)
	return svl, nil
}

func (s StateVersionService) CurrentStateVersion(ctx context.Context, workspaceID string) (*otf.StateVersion, error) {
	sv, err := s.db.GetStateVersion(ctx, otf.StateVersionGetOptions{WorkspaceID: &workspaceID})
	if err != nil {
		s.Error(err, "retrieving current state version", "workspace_id", workspaceID)
		return nil, err
	}
	s.V(2).Info("retrieved current state version", "workspace_id", workspaceID)
	return sv, nil
}

func (s StateVersionService) GetStateVersion(ctx context.Context, svID string) (*otf.StateVersion, error) {
	sv, err := s.db.GetStateVersion(ctx, otf.StateVersionGetOptions{ID: &svID})
	if err != nil {
		s.Error(err, "retrieving state version", "id", svID)
		return nil, err
	}
	s.V(2).Info("retrieved state version", "id", svID)
	return sv, nil
}

// Download state itself.
func (s StateVersionService) DownloadState(ctx context.Context, svID string) ([]byte, error) {
	if state, err := s.cache.Get(otf.StateVersionCacheKey(svID)); err == nil {
		s.V(2).Info("downloaded state", "id", svID)
		return state, nil
	}
	state, err := s.db.GetState(ctx, svID)
	if err != nil {
		s.Error(err, "downloading state", "id", svID)
		return nil, err
	}
	if err := s.cache.Set(otf.StateVersionCacheKey(svID), state); err != nil {
		return nil, fmt.Errorf("caching state: %w", err)
	}
	s.V(2).Info("downloaded state", "id", svID)
	return state, nil
}
