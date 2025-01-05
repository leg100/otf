package state

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/http/html"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/sql/sqlc"
	"github.com/leg100/otf/internal/tfeapi"
	"github.com/leg100/otf/internal/workspace"
	"github.com/leg100/surl/v2"
)

var ErrCurrentVersionDeletionAttempt = errors.New("deleting the current state version is not allowed")

// cacheKey generates a key for caching state files
func cacheKey(svID resource.ID) string { return fmt.Sprintf("%s.json", svID) }

type (
	// Service provides access to state and state versions
	Service struct {
		logr.Logger

		db     *pgdb
		cache  internal.Cache // cache state file
		web    *webHandlers
		tfeapi *tfe
		api    *api

		*factory // for creating state versions
		*authz.Authorizer
	}

	Options struct {
		logr.Logger
		html.Renderer
		internal.Cache
		*sql.DB
		*tfeapi.Responder
		*surl.Signer

		WorkspaceService *workspace.Service
		Authorizer       *authz.Authorizer
	}

	// StateVersionListOptions represents the options for listing state versions.
	StateVersionListOptions struct {
		resource.PageOptions
		Organization string `schema:"filter[organization][name],required"`
		Workspace    string `schema:"filter[workspace][name],required"`
	}
)

func NewService(opts Options) *Service {
	db := &pgdb{opts.DB}
	svc := Service{
		Logger:     opts.Logger,
		Authorizer: opts.Authorizer,
		cache:      opts.Cache,
		db:         db,
		factory:    &factory{db},
	}
	svc.web = &webHandlers{
		Renderer: opts.Renderer,
		Service:  &svc,
	}
	svc.tfeapi = &tfe{
		Responder:  opts.Responder,
		Signer:     opts.Signer,
		state:      &svc,
		workspaces: opts.WorkspaceService,
	}
	svc.api = &api{
		Service:   &svc,
		Responder: opts.Responder,
		tfeapi:    svc.tfeapi,
	}
	// include state version outputs in api responses when requested.
	opts.Responder.Register(tfeapi.IncludeOutputs, svc.tfeapi.includeOutputs)
	opts.Responder.Register(tfeapi.IncludeOutputs, svc.tfeapi.includeWorkspaceCurrentOutputs)

	// Resolve authorization requests for state version IDs to a workspace IDs
	opts.Authorizer.RegisterWorkspaceResolver(resource.StateVersionKind,
		func(ctx context.Context, svID resource.ID) (resource.ID, error) {
			sv, err := db.getVersion(ctx, svID)
			if err != nil {
				return resource.ID{}, err
			}
			return sv.WorkspaceID, nil
		},
	)
	return &svc
}

func (a *Service) AddHandlers(r *mux.Router) {
	a.web.addHandlers(r)
	a.tfeapi.addHandlers(r)
	a.api.addHandlers(r)
}

func (a *Service) Create(ctx context.Context, opts CreateStateVersionOptions) (*Version, error) {
	subject, err := a.Authorize(ctx, authz.CreateStateVersionAction, &authz.AccessRequest{ID: &opts.WorkspaceID})
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

func (a *Service) DownloadCurrent(ctx context.Context, workspaceID resource.ID) ([]byte, error) {
	v, err := a.GetCurrent(ctx, workspaceID)
	if err != nil {
		return nil, err
	}
	return a.Download(ctx, v.ID)
}

func (a *Service) List(ctx context.Context, workspaceID resource.ID, opts resource.PageOptions) (*resource.Page[*Version], error) {
	subject, err := a.Authorize(ctx, authz.ListStateVersionsAction, &authz.AccessRequest{ID: &workspaceID})
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

func (a *Service) GetCurrent(ctx context.Context, workspaceID resource.ID) (*Version, error) {
	subject, err := a.Authorize(ctx, authz.GetStateVersionAction, &authz.AccessRequest{ID: &workspaceID})
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

func (a *Service) Get(ctx context.Context, versionID resource.ID) (*Version, error) {
	subject, err := a.Authorize(ctx, authz.GetStateVersionAction, &authz.AccessRequest{ID: &versionID})
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

func (a *Service) Delete(ctx context.Context, versionID resource.ID) error {
	subject, err := a.Authorize(ctx, authz.DeleteStateVersionAction, &authz.AccessRequest{ID: &versionID})
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

func (a *Service) Rollback(ctx context.Context, versionID resource.ID) (*Version, error) {
	subject, err := a.Authorize(ctx, authz.RollbackStateVersionAction, &authz.AccessRequest{ID: &versionID})
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

func (a *Service) Upload(ctx context.Context, svID resource.ID, state []byte) error {
	var sv *Version
	err := a.db.Tx(ctx, func(ctx context.Context, q *sqlc.Queries) error {
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

func (a *Service) Download(ctx context.Context, svID resource.ID) ([]byte, error) {
	subject, err := a.Authorize(ctx, authz.DownloadStateAction, &authz.AccessRequest{ID: &svID})
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

func (a *Service) GetOutput(ctx context.Context, outputID resource.ID) (*Output, error) {
	out, err := a.db.getOutput(ctx, outputID)
	if err != nil {
		a.Error(err, "retrieving state version output", "id", outputID)
		return nil, err
	}

	subject, err := a.Authorize(ctx, authz.GetStateVersionOutputAction, &authz.AccessRequest{ID: &out.StateVersionID})
	if err != nil {
		return nil, err
	}

	a.V(9).Info("retrieved state version output", "id", outputID, "subject", subject)
	return out, nil
}
