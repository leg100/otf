package variable

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/leg100/otf/rbac"
	"github.com/pkg/errors"
)

type (
	service interface {
		create(ctx context.Context, workspaceID string, opts otf.CreateVariableOptions) (*otf.Variable, error)
		list(ctx context.Context, workspaceID string) ([]*otf.Variable, error)
		get(ctx context.Context, variableID string) (*otf.Variable, error)
		update(ctx context.Context, variableID string, opts otf.UpdateVariableOptions) (*otf.Variable, error)
		delete(ctx context.Context, variableID string) (*otf.Variable, error)
	}

	Service struct {
		logr.Logger

		db

		workspace otf.Authorizer

		api *api
		web *web
	}

	Options struct {
		WorkspaceAuthorizer otf.Authorizer
		otf.DB
		otf.Renderer
		otf.WorkspaceService
		logr.Logger
	}
)

func NewService(opts Options) *Service {
	svc := Service{
		Logger:    opts.Logger,
		workspace: opts.WorkspaceAuthorizer,
		db:        newPGDB(opts.DB),
	}

	svc.api = &api{
		svc: &svc,
	}
	svc.web = &web{
		svc:              &svc,
		Renderer:         opts.Renderer,
		WorkspaceService: opts.WorkspaceService,
	}

	return &svc
}

func (s *Service) AddHandlers(r *mux.Router) {
	s.api.addHandlers(r)
	s.web.addHandlers(r)
}

func (a *Service) ListVariables(ctx context.Context, workspaceID string) ([]*otf.Variable, error) {
	return a.list(ctx, workspaceID)
}

func (a *Service) create(ctx context.Context, workspaceID string, opts otf.CreateVariableOptions) (*otf.Variable, error) {
	subject, err := a.workspace.CanAccess(ctx, rbac.CreateVariableAction, workspaceID)
	if err != nil {
		return nil, err
	}

	v, err := otf.NewVariable(workspaceID, opts)
	if err != nil {
		a.Error(err, "constructing variable", "subject", subject, "workspace", workspaceID, "key", opts.Key)
		return nil, err
	}

	if err := a.db.create(ctx, v); err != nil {
		a.Error(err, "creating variable", "subject", subject, "variable", v)
		return nil, err
	}

	a.V(1).Info("created variable", "subject", subject, "variable", v)

	return v, nil
}

func (a *Service) get(ctx context.Context, variableID string) (*otf.Variable, error) {
	// retrieve variable first in order to retrieve workspace ID for authorization
	variable, err := a.db.get(ctx, variableID)
	if err != nil {
		a.Error(err, "retrieving variable", "workspace_id", variableID)
		return nil, err
	}

	subject, err := a.workspace.CanAccess(ctx, rbac.GetVariableAction, variable.WorkspaceID)
	if err != nil {
		return nil, err
	}

	a.V(2).Info("retrieved variable", "subject", subject, "variable", variable)

	return variable, nil
}

func (a *Service) update(ctx context.Context, variableID string, opts otf.UpdateVariableOptions) (*otf.Variable, error) {
	// retrieve existing in order to retrieve workspace ID for authorization
	existing, err := a.db.get(ctx, variableID)
	if err != nil {
		return nil, errors.Wrap(err, "retrieving variable")
	}

	subject, err := a.workspace.CanAccess(ctx, rbac.UpdateVariableAction, existing.WorkspaceID)
	if err != nil {
		return nil, err
	}

	updated, err := a.db.update(ctx, variableID, func(v *otf.Variable) error {
		return v.Update(opts)
	})
	if err != nil {
		a.Error(err, "updating variable", "subject", subject, "variable_id", variableID, "workspace_id", existing.WorkspaceID)
		return nil, err
	}
	a.V(1).Info("updated variable", "subject", subject, "before", existing, "after", updated)

	return updated, nil
}

func (a *Service) delete(ctx context.Context, variableID string) (*otf.Variable, error) {
	// retrieve existing in order to retrieve workspace ID for authorization
	existing, err := a.db.get(ctx, variableID)
	if err != nil {
		return nil, errors.Wrap(err, "retrieving variable")
	}

	subject, err := a.workspace.CanAccess(ctx, rbac.DeleteVariableAction, existing.WorkspaceID)
	if err != nil {
		return nil, err
	}

	deleted, err := a.db.delete(ctx, variableID)
	if err != nil {
		a.Error(err, "deleting variable", "subject", subject, "variable", existing)
		return nil, err
	}
	a.V(1).Info("deleted variable", "subject", subject, "variable", deleted)

	return deleted, nil
}

func (a *Service) list(ctx context.Context, workspaceID string) ([]*otf.Variable, error) {
	subject, err := a.workspace.CanAccess(ctx, rbac.ListVariablesAction, workspaceID)
	if err != nil {
		return nil, err
	}

	variables, err := a.db.list(ctx, workspaceID)
	if err != nil {
		a.Error(err, "listing variables", "subject", subject, "workspace_id", workspaceID)
		return nil, err
	}

	a.V(2).Info("listed variables", "subject", subject, "workspace_id", workspaceID)

	return variables, nil
}
