package variable

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
	"github.com/leg100/otf/rbac"
	"github.com/pkg/errors"
)

type application interface {
	create(ctx context.Context, workspaceID string, opts otf.CreateVariableOptions) (*Variable, error)
	list(ctx context.Context, workspaceID string) ([]*Variable, error)
	get(ctx context.Context, variableID string) (*Variable, error)
	update(ctx context.Context, variableID string, opts otf.UpdateVariableOptions) (*Variable, error)
	delete(ctx context.Context, variableID string) (*Variable, error)
}

type app struct {
	otf.WorkspaceAuthorizer
	logr.Logger

	db
}

func (a *app) ListVariables(ctx context.Context, workspaceID string) ([]otf.Variable, error) {
	vars, err := a.list(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	var iVars []otf.Variable
	for _, v := range vars {
		iVars = append(iVars, v)
	}
	return iVars, nil
}

func (a *app) create(ctx context.Context, workspaceID string, opts otf.CreateVariableOptions) (*Variable, error) {
	subject, err := a.CanAccessWorkspaceByID(ctx, rbac.CreateVariableAction, workspaceID)
	if err != nil {
		return nil, err
	}

	v, err := NewVariable(workspaceID, opts)
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

func (a *app) get(ctx context.Context, variableID string) (*Variable, error) {
	// retrieve variable first in order to retrieve workspace ID for authorization
	variable, err := a.db.get(ctx, variableID)
	if err != nil {
		a.Error(err, "retrieving variable", "workspace_id", variableID)
		return nil, err
	}

	subject, err := a.CanAccessWorkspaceByID(ctx, rbac.GetVariableAction, variable.WorkspaceID())
	if err != nil {
		return nil, err
	}

	a.V(2).Info("retrieved variable", "subject", subject, "variable", variable)

	return variable, nil
}

func (a *app) update(ctx context.Context, variableID string, opts otf.UpdateVariableOptions) (*Variable, error) {
	// retrieve existing in order to retrieve workspace ID for authorization
	existing, err := a.db.get(ctx, variableID)
	if err != nil {
		return nil, errors.Wrap(err, "retrieving variable")
	}

	subject, err := a.CanAccessWorkspaceByID(ctx, rbac.UpdateVariableAction, existing.WorkspaceID())
	if err != nil {
		return nil, err
	}

	updated, err := a.db.update(ctx, variableID, func(v *Variable) error {
		return v.Update(opts)
	})
	if err != nil {
		a.Error(err, "updating variable", "subject", subject, "variable_id", variableID, "workspace_id", existing.WorkspaceID())
		return nil, err
	}
	a.V(1).Info("updated variable", "subject", subject, "before", existing, "after", updated)

	return updated, nil
}

func (a *app) delete(ctx context.Context, variableID string) (*Variable, error) {
	// retrieve existing in order to retrieve workspace ID for authorization
	existing, err := a.db.get(ctx, variableID)
	if err != nil {
		return nil, errors.Wrap(err, "retrieving variable")
	}

	subject, err := a.CanAccessWorkspaceByID(ctx, rbac.DeleteVariableAction, existing.WorkspaceID())
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

func (a *app) list(ctx context.Context, workspaceID string) ([]*Variable, error) {
	subject, err := a.CanAccessWorkspaceByID(ctx, rbac.ListVariablesAction, workspaceID)
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
