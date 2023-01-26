package state

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
)

type Application struct {
	db    *pgdb     // access to state version database
	cache otf.Cache // cache state file
	logr.Logger
}

func (a *Application) CreateStateVersion(ctx context.Context, workspaceID string, opts otf.StateVersionCreateOptions) (*otf.StateVersion, error) {
	subject, err := a.CanAccessWorkspaceByID(ctx, otf.CreateStateVersionAction, workspaceID)
	if err != nil {
		return nil, err
	}

	sv, err := otf.NewStateVersion(opts)
	if err != nil {
		a.Error(err, "constructing state version")
		return nil, err
	}
	if err := a.db.CreateStateVersion(ctx, workspaceID, sv); err != nil {
		a.Error(err, "creating state version", "subject", subject)
		return nil, err
	}

	if err := a.cache.Set(otf.StateVersionCacheKey(sv.ID()), sv.State()); err != nil {
		return nil, fmt.Errorf("caching state version: %w", err)
	}

	a.V(0).Info("created state version", "id", sv.ID(), "workspace", workspaceID, "serial", sv.Serial(), "subject", subject)
	return sv, nil
}

func (a *Application) ListStateVersions(ctx context.Context, opts otf.StateVersionListOptions) (*otf.StateVersionList, error) {
	subject, err := a.CanAccessWorkspaceByName(ctx, otf.ListStateVersionsAction, opts.Organization, opts.Workspace)
	if err != nil {
		return nil, err
	}

	svl, err := a.db.ListStateVersions(ctx, opts)
	if err != nil {
		a.Error(err, "listing state versions", "organization", opts.Organization, "workspace", opts.Workspace, "subject", subject)
		return nil, err
	}
	a.V(2).Info("listed state version", "organization", opts.Organization, "workspace", opts.Workspace, "subject", subject)
	return svl, nil
}

func (a *Application) CurrentStateVersion(ctx context.Context, workspaceID string) (*otf.StateVersion, error) {
	subject, err := a.CanAccessWorkspaceByID(ctx, otf.GetStateVersionAction, workspaceID)
	if err != nil {
		return nil, err
	}

	sv, err := a.db.GetStateVersion(ctx, otf.StateVersionGetOptions{WorkspaceID: &workspaceID})
	if err != nil {
		a.Error(err, "retrieving current state version", "workspace_id", workspaceID, "subject", subject)
		return nil, err
	}
	a.V(2).Info("retrieved current state version", "workspace_id", workspaceID, "subject", subject)
	return sv, nil
}

func (a *Application) GetStateVersion(ctx context.Context, svID string) (*otf.StateVersion, error) {
	subject, err := a.CanAccessStateVersion(ctx, otf.GetStateVersionAction, svID)
	if err != nil {
		return nil, err
	}

	sv, err := a.db.GetStateVersion(ctx, otf.StateVersionGetOptions{ID: &svID})
	if err != nil {
		a.Error(err, "retrieving state version", "id", svID, "subject", subject)
		return nil, err
	}
	a.V(2).Info("retrieved state version", "id", svID, "subject", subject)
	return sv, nil
}

// DownloadState retrieves base64-encoded terraform state from the db
func (a *Application) DownloadState(ctx context.Context, svID string) ([]byte, error) {
	subject, err := a.CanAccessStateVersion(ctx, otf.DownloadStateAction, svID)
	if err != nil {
		return nil, err
	}

	if state, err := a.cache.Get(otf.StateVersionCacheKey(svID)); err == nil {
		a.V(2).Info("downloaded state", "id", svID, "subject", subject)
		return state, nil
	}
	state, err := a.db.GetState(ctx, svID)
	if err != nil {
		a.Error(err, "downloading state", "id", svID, "subject", subject)
		return nil, err
	}
	if err := a.cache.Set(otf.StateVersionCacheKey(svID), state); err != nil {
		return nil, fmt.Errorf("caching state: %w", err)
	}
	a.V(2).Info("downloaded state", "id", svID, "subject", subject)
	return state, nil
}
