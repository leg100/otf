package state

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
	"github.com/leg100/otf/rbac"
)

// cacheKey generates a key for caching state files
func cacheKey(svID string) string { return fmt.Sprintf("%s.json", svID) }

var _ otf.StateVersionService = (*Service)(nil)

// service is the application service for state
type (
	service interface {
		createVersion(ctx context.Context, opts otf.CreateStateVersionOptions) (*version, error)
		currentVersion(ctx context.Context, workspaceID string) (*version, error)
		getVersion(ctx context.Context, versionID string) (*version, error)
		downloadState(ctx context.Context, versionID string) ([]byte, error)
		listVersions(ctx context.Context, opts stateVersionListOptions) (*versionList, error)
		getOutput(ctx context.Context, outputID string) (*output, error)
	}

	// Service provides access to state and state versions
	Service struct {
		logr.Logger
		otf.WorkspaceService

		db                  // access to state version database
		cache     otf.Cache // cache state file
		workspace otf.Authorizer

		*api
	}
	Options struct {
		logr.Logger

		WorkspaceAuthorizer otf.Authorizer

		otf.Cache
		otf.DB
	}

	// stateVersionGetOptions are options for retrieving a single StateVersion.
	// Either ID *or* WorkspaceID must be specfiied.
	stateVersionGetOptions struct {
		// ID of state version to retrieve
		ID *string
		// Get current state version belonging to workspace with this ID
		WorkspaceID *string
	}

	// stateVersionListOptions represents the options for listing state versions.
	stateVersionListOptions struct {
		otf.ListOptions
		Organization string `schema:"filter[organization][name],required"`
		Workspace    string `schema:"filter[workspace][name],required"`
	}
)

func NewService(opts Options) *Service {
	svc := Service{
		Logger: opts.Logger,
	}

	svc.db = newPGDB(opts.DB)
	svc.cache = opts.Cache
	svc.workspace = opts.WorkspaceAuthorizer

	svc.api = &api{&svc}

	return &svc
}

func (a *Service) CreateStateVersion(ctx context.Context, opts otf.CreateStateVersionOptions) error {
	_, err := a.createVersion(ctx, opts)
	return err
}

func (a *Service) DownloadCurrentState(ctx context.Context, workspaceID string) ([]byte, error) {
	v, err := a.currentVersion(ctx, workspaceID)
	if err != nil {
		return nil, err
	}
	return a.downloadState(ctx, v.ID)
}

func (a *Service) createVersion(ctx context.Context, opts otf.CreateStateVersionOptions) (*version, error) {
	if opts.WorkspaceID == nil {
		return nil, errors.New("workspace ID is required")
	}
	subject, err := a.workspace.CanAccess(ctx, rbac.CreateStateVersionAction, *opts.WorkspaceID)
	if err != nil {
		return nil, err
	}

	sv, err := newVersion(opts)
	if err != nil {
		a.Error(err, "constructing state version")
		return nil, err
	}
	if err := a.db.createVersion(ctx, sv); err != nil {
		a.Error(err, "creating state version", "subject", subject)
		return nil, err
	}

	if err := a.cache.Set(cacheKey(sv.ID), sv.State); err != nil {
		return nil, fmt.Errorf("caching state version: %w", err)
	}

	a.V(0).Info("created state version", "id", sv.ID, "workspace", *opts.WorkspaceID, "serial", sv.Serial, "subject", subject)
	return sv, nil
}

func (a *Service) listVersions(ctx context.Context, opts stateVersionListOptions) (*versionList, error) {
	workspace, err := a.GetWorkspaceByName(ctx, opts.Organization, opts.Workspace)
	if err != nil {
		return nil, err
	}
	subject, err := a.workspace.CanAccess(ctx, rbac.ListStateVersionsAction, workspace.ID)
	if err != nil {
		return nil, err
	}

	svl, err := a.db.listVersions(ctx, opts)
	if err != nil {
		a.Error(err, "listing state versions", "organization", opts.Organization, "workspace", opts.Workspace, "subject", subject)
		return nil, err
	}
	a.V(2).Info("listed state version", "organization", opts.Organization, "workspace", opts.Workspace, "subject", subject)
	return svl, nil
}

func (a *Service) currentVersion(ctx context.Context, workspaceID string) (*version, error) {
	subject, err := a.workspace.CanAccess(ctx, rbac.GetStateVersionAction, workspaceID)
	if err != nil {
		return nil, err
	}

	sv, err := a.db.getVersion(ctx, stateVersionGetOptions{WorkspaceID: &workspaceID})
	if err != nil {
		a.Error(err, "retrieving current state version", "workspace_id", workspaceID, "subject", subject)
		return nil, err
	}
	a.V(2).Info("retrieved current state version", "workspace_id", workspaceID, "subject", subject)
	return sv, nil
}

func (a *Service) getVersion(ctx context.Context, versionID string) (*version, error) {
	subject, err := a.CanAccessStateVersion(ctx, rbac.GetStateVersionAction, versionID)
	if err != nil {
		return nil, err
	}

	sv, err := a.db.getVersion(ctx, stateVersionGetOptions{ID: &versionID})
	if err != nil {
		a.Error(err, "retrieving state version", "id", versionID, "subject", subject)
		return nil, err
	}
	a.V(2).Info("retrieved state version", "id", versionID, "subject", subject)
	return sv, nil
}

// DownloadState retrieves base64-encoded terraform state from the db
func (a *Service) downloadState(ctx context.Context, svID string) ([]byte, error) {
	subject, err := a.CanAccessStateVersion(ctx, rbac.DownloadStateAction, svID)
	if err != nil {
		return nil, err
	}

	if state, err := a.cache.Get(cacheKey(svID)); err == nil {
		a.V(2).Info("downloaded state", "id", svID, "subject", subject)
		return state, nil
	}
	state, err := a.db.getState(ctx, svID)
	if err != nil {
		a.Error(err, "downloading state", "id", svID, "subject", subject)
		return nil, err
	}
	if err := a.cache.Set(cacheKey(svID), state); err != nil {
		return nil, fmt.Errorf("caching state: %w", err)
	}
	a.V(2).Info("downloaded state", "id", svID, "subject", subject)
	return state, nil
}

func (a *Service) getOutput(ctx context.Context, outputID string) (*output, error) {
	sv, err := a.db.getOutput(ctx, outputID)
	if err != nil {
		a.Error(err, "retrieving state version output", "id", outputID)
		return nil, err
	}

	subject, err := a.CanAccessStateVersion(ctx, rbac.GetStateVersionOutputAction, sv.stateVersionID)
	if err != nil {
		return nil, err
	}

	a.V(2).Info("retrieved state version output", "id", outputID, "subject", subject)
	return sv, nil
}

func (a *Service) CanAccessStateVersion(ctx context.Context, action rbac.Action, svID string) (otf.Subject, error) {
	sv, err := a.db.getVersion(ctx, stateVersionGetOptions{ID: &svID})
	if err != nil {
		return nil, err
	}
	return a.workspace.CanAccess(ctx, action, sv.WorkspaceID)
}
