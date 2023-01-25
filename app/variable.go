package app

import (
	"context"

	"github.com/leg100/otf"
	"github.com/leg100/otf/rbac"
	"github.com/pkg/errors"
)

func (a *Application) CreateVariable(ctx context.Context, workspaceID string, opts otf.CreateVariableOptions) (*otf.Variable, error) {
	subject, err := a.CanAccessWorkspaceByID(ctx, rbac.CreateVariableAction, workspaceID)
	if err != nil {
		return nil, err
	}

	v, err := otf.NewVariable(workspaceID, opts)
	if err != nil {
		a.Error(err, "constructing variable", "subject", subject, "workspace", workspaceID, "key", opts.Key)
		return nil, err
	}

	if err := a.db.CreateVariable(ctx, v); err != nil {
		a.Error(err, "creating variable", "subject", subject, "variable", v)
		return nil, err
	}

	a.V(1).Info("created variable", "subject", subject, "variable", v)

	return v, nil
}

func (a *Application) ListVariables(ctx context.Context, workspaceID string) ([]*otf.Variable, error) {
	subject, err := a.CanAccessWorkspaceByID(ctx, rbac.ListVariablesAction, workspaceID)
	if err != nil {
		return nil, err
	}

	variables, err := a.db.ListVariables(ctx, workspaceID)
	if err != nil {
		a.Error(err, "listing variables", "subject", subject, "workspace_id", workspaceID)
		return nil, err
	}

	a.V(2).Info("listed variables", "subject", subject, "workspace_id", workspaceID)

	return variables, nil
}

func (a *Application) GetVariable(ctx context.Context, variableID string) (*otf.Variable, error) {
	// retrieve variable first in order to retrieve workspace ID for authorization
	variable, err := a.db.GetVariable(ctx, variableID)
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

func (a *Application) UpdateVariable(ctx context.Context, variableID string, opts otf.UpdateVariableOptions) (*otf.Variable, error) {
	// retrieve existing in order to retrieve workspace ID for authorization
	existing, err := a.db.GetVariable(ctx, variableID)
	if err != nil {
		return nil, errors.Wrap(err, "retrieving variable")
	}

	subject, err := a.CanAccessWorkspaceByID(ctx, rbac.UpdateVariableAction, existing.WorkspaceID())
	if err != nil {
		return nil, err
	}

	updated, err := a.db.UpdateVariable(ctx, variableID, func(v *otf.Variable) error {
		return v.Update(opts)
	})
	if err != nil {
		a.Error(err, "updating variable", "subject", subject, "variable_id", variableID, "workspace_id", existing.WorkspaceID())
		return nil, err
	}
	a.V(1).Info("updated variable", "subject", subject, "before", existing, "after", updated)

	return updated, nil
}

func (a *Application) DeleteVariable(ctx context.Context, variableID string) (*otf.Variable, error) {
	// retrieve existing in order to retrieve workspace ID for authorization
	existing, err := a.db.GetVariable(ctx, variableID)
	if err != nil {
		return nil, errors.Wrap(err, "retrieving variable")
	}

	subject, err := a.CanAccessWorkspaceByID(ctx, rbac.DeleteVariableAction, existing.WorkspaceID())
	if err != nil {
		return nil, err
	}

	deleted, err := a.db.DeleteVariable(ctx, variableID)
	if err != nil {
		a.Error(err, "deleting variable", "subject", subject, "variable", existing)
		return nil, err
	}
	a.V(1).Info("deleted variable", "subject", subject, "variable", deleted)

	return deleted, nil
}
