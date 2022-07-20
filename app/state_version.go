package app

import (
	"context"
	"fmt"

	"github.com/leg100/otf"
)

func (a *Application) CreateStateVersion(ctx context.Context, workspaceID string, opts otf.StateVersionCreateOptions) (*otf.StateVersion, error) {
	sv, err := otf.NewStateVersion(opts)
	if err != nil {
		a.Error(err, "constructing state version")
		return nil, err
	}
	if err := a.db.CreateStateVersion(ctx, workspaceID, sv); err != nil {
		a.Error(err, "creating state version")
		return nil, err
	}

	if err := a.cache.Set(otf.StateVersionCacheKey(sv.ID()), sv.State()); err != nil {
		return nil, fmt.Errorf("caching state version: %w", err)
	}

	a.V(0).Info("created state version", "id", sv.ID(), "workspace", workspaceID, "serial", sv.Serial())
	return sv, nil
}

func (a *Application) ListStateVersion(ctx context.Context, opts otf.StateVersionListOptions) (*otf.StateVersionList, error) {
	svl, err := a.db.ListStateVersions(ctx, opts)
	if err != nil {
		a.Error(err, "listing state versions", opts.LogFields()...)
		return nil, err
	}
	a.V(2).Info("listed state version", opts.LogFields()...)
	return svl, nil
}

func (a *Application) CurrentStateVersion(ctx context.Context, workspaceID string) (*otf.StateVersion, error) {
	sv, err := a.db.GetStateVersion(ctx, otf.StateVersionGetOptions{WorkspaceID: &workspaceID})
	if err != nil {
		a.Error(err, "retrieving current state version", "workspace_id", workspaceID)
		return nil, err
	}
	a.V(2).Info("retrieved current state version", "workspace_id", workspaceID)
	return sv, nil
}

func (a *Application) GetStateVersion(ctx context.Context, svID string) (*otf.StateVersion, error) {
	sv, err := a.db.GetStateVersion(ctx, otf.StateVersionGetOptions{ID: &svID})
	if err != nil {
		a.Error(err, "retrieving state version", "id", svID)
		return nil, err
	}
	a.V(2).Info("retrieved state version", "id", svID)
	return sv, nil
}

// DownloadState retrieves base64-encoded terraform state from the db
func (a *Application) DownloadState(ctx context.Context, svID string) ([]byte, error) {
	if state, err := a.cache.Get(otf.StateVersionCacheKey(svID)); err == nil {
		a.V(2).Info("downloaded state", "id", svID)
		return state, nil
	}
	state, err := a.db.GetState(ctx, svID)
	if err != nil {
		a.Error(err, "downloading state", "id", svID)
		return nil, err
	}
	if err := a.cache.Set(otf.StateVersionCacheKey(svID), state); err != nil {
		return nil, fmt.Errorf("caching state: %w", err)
	}
	a.V(2).Info("downloaded state", "id", svID)
	return state, nil
}
