package state

import (
	"context"
	"errors"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/sql"
)

var ErrCurrentVersionDeletionAttempt = errors.New("deleting the current state version is not allowed")

type (
	// Alias service to permit embedding it with other services in a struct
	// without a name clash.
	StateService = Service

	// Service provides access to state and state versions
	Service struct {
		logr.Logger

		db *pgdb

		*factory // for creating state versions
		*authz.Authorizer
	}

	Options struct {
		Logger     logr.Logger
		DB         *sql.DB
		Authorizer *authz.Authorizer
	}
)

func NewService(opts Options) *Service {
	db := &pgdb{opts.DB}
	svc := Service{
		Logger:     opts.Logger,
		Authorizer: opts.Authorizer,
		db:         db,
		factory:    &factory{db},
	}

	// Provide a means of looking up a state versions's parent workspace.
	opts.Authorizer.RegisterParentResolver(resource.StateVersionKind,
		func(ctx context.Context, svID resource.ID) (resource.ID, error) {
			// NOTE: we look up directly in the database rather than via
			// service call to avoid a recursion loop.
			sv, err := db.getVersion(ctx, svID)
			if err != nil {
				return nil, err
			}
			return sv.WorkspaceID, nil
		},
	)
	return &svc
}

func (a *Service) CreateStateVersion(ctx context.Context, opts CreateStateVersionOptions) (*Version, error) {
	subject, err := a.Authorize(ctx, authz.CreateStateVersionAction, opts.WorkspaceID)
	if err != nil {
		return nil, err
	}

	sv, err := a.new(ctx, opts)
	if err != nil {
		a.Error(err, "creating state version", "subject", subject)
		return nil, err
	}

	a.V(0).Info("created state version", "state_version", sv, "subject", subject)
	return sv, nil
}

func (a *Service) DownloadCurrentState(ctx context.Context, workspaceID resource.TfeID) ([]byte, error) {
	v, err := a.GetCurrentStateVersion(ctx, workspaceID)
	if err != nil {
		return nil, err
	}
	return a.DownloadState(ctx, v.ID)
}

func (a *Service) ListStateVersions(ctx context.Context, workspaceID resource.TfeID, opts resource.PageOptions) (*resource.Page[*Version], error) {
	subject, err := a.Authorize(ctx, authz.ListStateVersionsAction, workspaceID)
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

func (a *Service) GetCurrentStateVersion(ctx context.Context, workspaceID resource.TfeID) (*Version, error) {
	subject, err := a.Authorize(ctx, authz.GetStateVersionAction, workspaceID)
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

// GetPreviousStateVersion returns the finalized state version that immediately precedes sv
// (by serial) in the same workspace. Returns ErrResourceNotFound when sv is
// the first version.
func (a *Service) GetPreviousStateVersion(ctx context.Context, sv *Version) (*Version, error) {
	if _, err := a.Authorize(ctx, authz.GetStateVersionAction, sv.WorkspaceID); err != nil {
		return nil, err
	}
	prev, err := a.db.getPreviousVersion(ctx, sv)
	if err != nil {
		return nil, err
	}
	a.V(9).Info("retrieved previous state version", "state_version", prev)
	return prev, nil
}

func (a *Service) GetStateVersion(ctx context.Context, versionID resource.TfeID) (*Version, error) {
	subject, err := a.Authorize(ctx, authz.GetStateVersionAction, versionID)
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

func (a *Service) DeleteStateVersion(ctx context.Context, versionID resource.TfeID) error {
	subject, err := a.Authorize(ctx, authz.DeleteStateVersionAction, versionID)
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

func (a *Service) RollbackStateVersion(ctx context.Context, versionID resource.TfeID) (*Version, error) {
	subject, err := a.Authorize(ctx, authz.RollbackStateVersionAction, versionID)
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

func (a *Service) UploadState(ctx context.Context, svID resource.TfeID, state []byte) error {
	var sv *Version
	err := a.db.Tx(ctx, func(ctx context.Context) error {
		var err error
		sv, err = a.db.getVersionForUpdate(ctx, svID)
		if err != nil {
			return err
		}
		sv, err = a.uploadStateAndOutputs(ctx, sv, state)
		if err != nil {
			return err
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

func (a *Service) DownloadState(ctx context.Context, svID resource.TfeID) ([]byte, error) {
	subject, err := a.Authorize(ctx, authz.DownloadStateAction, svID)
	if err != nil {
		return nil, err
	}
	state, err := a.db.getState(ctx, svID)
	if err != nil {
		a.Error(err, "downloading state", "id", svID, "subject", subject)
		return nil, err
	}
	a.V(9).Info("downloaded state", "id", svID, "subject", subject)
	return state, nil
}

func (a *Service) GetStateOutput(ctx context.Context, outputID resource.TfeID) (*Output, error) {
	out, err := a.db.getOutput(ctx, outputID)
	if err != nil {
		a.Error(err, "retrieving state version output", "id", outputID)
		return nil, err
	}

	subject, err := a.Authorize(ctx, authz.GetStateVersionOutputAction, out.StateVersionID)
	if err != nil {
		return nil, err
	}

	a.V(9).Info("retrieved state version output", "id", outputID, "subject", subject)
	return out, nil
}
