package state

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
	"github.com/leg100/otf/rbac"
)

type Application interface {
	CreateStateVersion(ctx context.Context, opts otf.CreateStateVersionOptions) (*Version, error)
	CurrentStateVersion(ctx context.Context, workspaceID string) (*Version, error)
	GetStateVersion(ctx context.Context, id string) (*Version, error)
	DownloadState(ctx context.Context, id string) ([]byte, error)
	ListStateVersions(ctx context.Context, opts StateVersionListOptions) (*StateVersionList, error)
}

type app struct {
	otf.Authorizer // authorize access
	logr.Logger

	db    *pgdb     // access to state version database
	cache otf.Cache // cache state file
}

func NewApp(opts AppOptions) *app {
	return &app{
		Authorizer: opts.Authorizer,
		Logger:     opts.Logger,
		db:         newPGDB(opts.Database),
		cache:      opts.Cache,
	}
}

type AppOptions struct {
	otf.Authorizer
	otf.Database
	otf.Cache
	logr.Logger
}

func (a *app) CreateStateVersion(ctx context.Context, opts otf.CreateStateVersionOptions) (*Version, error) {
	if opts.WorkspaceID == nil {
		return nil, errors.New("workspace ID is required")
	}
	subject, err := a.CanAccessWorkspaceByID(ctx, rbac.CreateStateVersionAction, *opts.WorkspaceID)
	if err != nil {
		return nil, err
	}

	sv, err := NewStateVersion(opts)
	if err != nil {
		a.Error(err, "constructing state version")
		return nil, err
	}
	if err := a.db.CreateStateVersion(ctx, sv); err != nil {
		a.Error(err, "creating state version", "subject", subject)
		return nil, err
	}

	if err := a.cache.Set(otf.StateVersionCacheKey(sv.ID()), sv.State()); err != nil {
		return nil, fmt.Errorf("caching state version: %w", err)
	}

	a.V(0).Info("created state version", "id", sv.ID(), "workspace", *opts.WorkspaceID, "serial", sv.Serial(), "subject", subject)
	return sv, nil
}

func (a *app) ListStateVersions(ctx context.Context, opts otf.StateVersionListOptions) (*StateVersionList, error) {
	subject, err := a.CanAccessWorkspaceByName(ctx, rbac.ListStateVersionsAction, opts.Organization, opts.Workspace)
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

func (a *app) CurrentStateVersion(ctx context.Context, workspaceID string) (*Version, error) {
	subject, err := a.CanAccessWorkspaceByID(ctx, rbac.GetStateVersionAction, workspaceID)
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

func (a *app) GetStateVersion(ctx context.Context, svID string) (*Version, error) {
	subject, err := a.CanAccessStateVersion(ctx, rbac.GetStateVersionAction, svID)
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
func (a *app) DownloadState(ctx context.Context, svID string) ([]byte, error) {
	subject, err := a.CanAccessStateVersion(ctx, rbac.DownloadStateAction, svID)
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
