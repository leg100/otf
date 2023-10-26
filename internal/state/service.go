package state

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/http/html"
	"github.com/leg100/otf/internal/rbac"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/sql/pggen"
	"github.com/leg100/otf/internal/tfeapi"
	"github.com/leg100/otf/internal/workspace"
	"github.com/leg100/surl"
)

var ErrCurrentVersionDeletionAttempt = errors.New("deleting the current state version is not allowed")

// cacheKey generates a key for caching state files
func cacheKey(svID string) string { return fmt.Sprintf("%s.json", svID) }

type (
	// Alias services so they don't conflict when nested together in struct
	StateService = Service

	// Service is the application Service for state
	Service interface {
		// CreateStateVersion creates a state version for a workspace,
		// optionally including the state data, or uploading it later using
		// UploadStateVersion.
		CreateStateVersion(ctx context.Context, opts CreateStateVersionOptions) (*Version, error)
		// DownloadCurrentState downloads the current (latest) state for the given
		// workspace.
		DownloadCurrentState(ctx context.Context, workspaceID string) ([]byte, error)
		ListStateVersions(ctx context.Context, workspaceID string, opts resource.PageOptions) (*resource.Page[*Version], error)
		GetCurrentStateVersion(ctx context.Context, workspaceID string) (*Version, error)
		GetStateVersion(ctx context.Context, versionID string) (*Version, error)
		DeleteStateVersion(ctx context.Context, versionID string) error
		// RollbackStateVersion creates a state version by duplicating the
		// specified state version and sets it as the current state version for
		// the given workspace.
		RollbackStateVersion(ctx context.Context, versionID string) (*Version, error)
		// UploadState uploads the state data for a state version.
		UploadState(ctx context.Context, versionID string, state []byte) error
		// DownloadState downloads the state data for a state version.
		DownloadState(ctx context.Context, versionID string) ([]byte, error)
		GetStateVersionOutput(ctx context.Context, outputID string) (*Output, error)
	}

	// service provides access to state and state versions
	service struct {
		logr.Logger

		db        *pgdb
		cache     internal.Cache // cache state file
		workspace internal.Authorizer
		web       *webHandlers
		tfeapi    *tfe
		api       *api

		*factory // for creating state versions
	}

	Options struct {
		logr.Logger
		html.Renderer

		WorkspaceAuthorizer internal.Authorizer

		internal.Cache
		workspace.WorkspaceService
		*sql.DB
		*tfeapi.Responder
		*surl.Signer
	}

	// StateVersionListOptions represents the options for listing state versions.
	StateVersionListOptions struct {
		resource.PageOptions
		Organization string `schema:"filter[organization][name],required"`
		Workspace    string `schema:"filter[workspace][name],required"`
	}
)

func NewService(opts Options) *service {
	db := &pgdb{opts.DB}
	svc := service{
		Logger:    opts.Logger,
		cache:     opts.Cache,
		db:        db,
		workspace: opts.WorkspaceAuthorizer,
		factory:   &factory{db},
	}
	svc.web = &webHandlers{
		Renderer: opts.Renderer,
		Service:  &svc,
	}
	svc.api = &api{
		Service:   &svc,
		Responder: opts.Responder,
	}
	svc.tfeapi = &tfe{
		Service:          &svc,
		WorkspaceService: opts.WorkspaceService,
		Responder:        opts.Responder,
		Signer:           opts.Signer,
	}
	// include state version outputs in api responses when requested.
	opts.Responder.Register(tfeapi.IncludeOutputs, svc.tfeapi.includeOutputs)
	opts.Responder.Register(tfeapi.IncludeOutputs, svc.tfeapi.includeWorkspaceCurrentOutputs)
	return &svc
}

func (a *service) AddHandlers(r *mux.Router) {
	a.web.addHandlers(r)
	a.tfeapi.addHandlers(r)
	a.api.addHandlers(r)
}

func (a *service) CreateStateVersion(ctx context.Context, opts CreateStateVersionOptions) (*Version, error) {
	if opts.WorkspaceID == nil {
		return nil, errors.New("workspace ID is required")
	}
	subject, err := a.workspace.CanAccess(ctx, rbac.CreateStateVersionAction, *opts.WorkspaceID)
	if err != nil {
		return nil, err
	}

	sv, err := a.new(ctx, opts)
	if err != nil {
		a.Error(err, "creating state version", "subject", subject)
		return nil, err
	}

	if err := a.cache.Set(cacheKey(sv.ID), sv.State); err != nil {
		a.Error(err, "caching state file")
	}

	a.V(0).Info("created state version", "state_version", sv, "subject", subject)
	return sv, nil
}

func (a *service) DownloadCurrentState(ctx context.Context, workspaceID string) ([]byte, error) {
	v, err := a.GetCurrentStateVersion(ctx, workspaceID)
	if err != nil {
		return nil, err
	}
	return a.DownloadState(ctx, v.ID)
}

func (a *service) ListStateVersions(ctx context.Context, workspaceID string, opts resource.PageOptions) (*resource.Page[*Version], error) {
	subject, err := a.workspace.CanAccess(ctx, rbac.ListStateVersionsAction, workspaceID)
	if err != nil {
		return nil, err
	}

	svl, err := a.db.listVersions(ctx, workspaceID, opts)
	if err != nil {
		a.Error(err, "listing state versions", "workspace", workspaceID, "subject", subject)
		return nil, err
	}
	a.V(9).Info("listed state versions", "workspace", workspaceID, "subject", subject)
	return svl, nil
}

func (a *service) GetCurrentStateVersion(ctx context.Context, workspaceID string) (*Version, error) {
	subject, err := a.workspace.CanAccess(ctx, rbac.GetStateVersionAction, workspaceID)
	if err != nil {
		return nil, err
	}

	sv, err := a.db.getCurrentVersion(ctx, workspaceID)
	if errors.Is(err, internal.ErrResourceNotFound) {
		// not found error occurs legitimately with a new workspace without any
		// state, so we log these errors at low level instead
		a.V(3).Info("retrieving current state version: workspace has no state yet", "workspace_id", workspaceID, "subject", subject)
		return nil, err
	} else if err != nil {
		a.Error(err, "retrieving current state version", "workspace_id", workspaceID, "subject", subject)
		return nil, err
	}
	a.V(9).Info("retrieved current state version", "state_version", sv, "subject", subject)
	return sv, nil
}

func (a *service) GetStateVersion(ctx context.Context, versionID string) (*Version, error) {
	subject, err := a.CanAccessStateVersion(ctx, rbac.GetStateVersionAction, versionID)
	if err != nil {
		return nil, err
	}

	sv, err := a.db.getVersion(ctx, versionID)
	if err != nil {
		a.Error(err, "retrieving state version", "id", versionID, "subject", subject)
		return nil, err
	}
	a.V(9).Info("retrieved state version", "state_version", sv, "subject", subject)
	return sv, nil
}

func (a *service) DeleteStateVersion(ctx context.Context, versionID string) error {
	subject, err := a.CanAccessStateVersion(ctx, rbac.DeleteStateVersionAction, versionID)
	if err != nil {
		return err
	}

	if err := a.db.deleteVersion(ctx, versionID); err != nil {
		a.Error(err, "deleting state version", "id", versionID, "subject", subject)
		return err
	}
	a.V(0).Info("deleted state version", "id", versionID, "subject", subject)
	return nil
}

func (a *service) RollbackStateVersion(ctx context.Context, versionID string) (*Version, error) {
	subject, err := a.CanAccessStateVersion(ctx, rbac.RollbackStateVersionAction, versionID)
	if err != nil {
		return nil, err
	}

	sv, err := a.rollback(ctx, versionID)
	if err != nil {
		a.Error(err, "rolling back state version", "id", versionID, "subject", subject)
		return nil, err
	}
	a.V(0).Info("rolled back state version", "state_version", sv, "subject", subject)
	return sv, nil
}

// NOTE: unauthenticated - access granted only via signed URL
func (a *service) UploadState(ctx context.Context, svID string, state []byte) error {
	var sv *Version
	err := a.db.Tx(ctx, func(ctx context.Context, q pggen.Querier) error {
		var err error
		sv, err = a.db.getVersionForUpdate(ctx, svID)
		if err != nil {
			return err
		}
		sv, err = a.uploadStateAndOutputs(ctx, sv, state)
		if err != nil {
			return err
		}
		if err := a.cache.Set(cacheKey(svID), state); err != nil {
			a.Error(err, "caching state file")
		}
		return nil
	})
	if err != nil {
		a.Error(err, "uploading state", "id", svID)
		return err
	}
	a.V(9).Info("uploading state", "state_version", sv)
	return nil
}

func (a *service) DownloadState(ctx context.Context, svID string) ([]byte, error) {
	subject, err := a.CanAccessStateVersion(ctx, rbac.DownloadStateAction, svID)
	if err != nil {
		return nil, err
	}
	if state, err := a.cache.Get(cacheKey(svID)); err == nil {
		a.V(9).Info("downloaded state", "id", svID, "subject", subject)
		return state, nil
	}
	state, err := a.db.getState(ctx, svID)
	if err != nil {
		a.Error(err, "downloading state", "id", svID, "subject", subject)
		return nil, err
	}
	if err := a.cache.Set(cacheKey(svID), state); err != nil {
		a.Error(err, "caching state file")
	}
	a.V(9).Info("downloaded state", "id", svID, "subject", subject)
	return state, nil
}

func (a *service) GetStateVersionOutput(ctx context.Context, outputID string) (*Output, error) {
	out, err := a.db.getOutput(ctx, outputID)
	if err != nil {
		a.Error(err, "retrieving state version output", "id", outputID)
		return nil, err
	}

	subject, err := a.CanAccessStateVersion(ctx, rbac.GetStateVersionOutputAction, out.StateVersionID)
	if err != nil {
		return nil, err
	}

	a.V(9).Info("retrieved state version output", "id", outputID, "subject", subject)
	return out, nil
}

func (a *service) CanAccessStateVersion(ctx context.Context, action rbac.Action, svID string) (internal.Subject, error) {
	sv, err := a.db.getVersion(ctx, svID)
	if err != nil {
		return nil, err
	}
	return a.workspace.CanAccess(ctx, action, sv.WorkspaceID)
}
