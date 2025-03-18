package variable

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/run"
	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/tfeapi"
	"github.com/leg100/otf/internal/workspace"
)

type (
	Service struct {
		logr.Logger
		*authz.Authorizer

		db     *pgdb
		web    *web
		tfeapi *tfe
		api    *api
		runs   runClient
	}

	Options struct {
		WorkspaceService *workspace.Service
		RunClient        runClient
		Authorizer       *authz.Authorizer

		*sql.DB
		*tfeapi.Responder
		logr.Logger
	}

	runClient interface {
		Get(ctx context.Context, runID resource.ID) (*run.Run, error)
	}
)

func NewService(opts Options) *Service {
	svc := Service{
		Logger:     opts.Logger,
		Authorizer: opts.Authorizer,
		db:         &pgdb{opts.DB},
		runs:       opts.RunClient,
	}

	svc.web = &web{
		authorizer: opts.Authorizer,
		workspaces: opts.WorkspaceService,
		variables:  &svc,
	}
	svc.tfeapi = &tfe{
		Service:   &svc,
		Responder: opts.Responder,
	}
	svc.api = &api{
		Service:   &svc,
		Responder: opts.Responder,
	}

	return &svc
}

func (s *Service) AddHandlers(r *mux.Router) {
	s.web.addHandlers(r)
	s.tfeapi.addHandlers(r)
	s.api.addHandlers(r)
}

func (s *Service) ListEffectiveVariables(ctx context.Context, runID resource.ID) ([]*Variable, error) {
	run, err := s.runs.Get(ctx, runID)
	if err != nil {
		return nil, err
	}
	sets, err := s.listWorkspaceVariableSets(ctx, run.WorkspaceID)
	if err != nil {
		return nil, err
	}
	vars, err := s.ListWorkspaceVariables(ctx, run.WorkspaceID)
	if err != nil {
		return nil, err
	}
	return mergeVariables(sets, vars, run), nil
}

func (s *Service) CreateWorkspaceVariable(ctx context.Context, workspaceID resource.ID, opts CreateVariableOptions) (*Variable, error) {
	subject, err := s.Authorize(ctx, authz.CreateWorkspaceVariableAction, &authz.AccessRequest{ID: &workspaceID})
	if err != nil {
		return nil, err
	}

	var v *Variable
	err = s.db.Lock(ctx, "variables", func(ctx context.Context, _ sql.Connection) (err error) {
		workspaceVars, err := s.ListWorkspaceVariables(ctx, workspaceID)
		if err != nil {
			return err
		}

		v, err = newVariable(workspaceVars, opts)
		if err != nil {
			return err
		}

		if err := s.db.createWorkspaceVariable(ctx, workspaceID, v); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		s.Error(err, "creating workspace variable", "subject", subject, "workspace_id", workspaceID, "variable", v)
		return nil, err
	}

	s.V(1).Info("created workspace variable", "subject", subject, "workspace_id", workspaceID, "variable", v)

	return v, nil
}

func (s *Service) UpdateWorkspaceVariable(ctx context.Context, variableID resource.ID, opts UpdateVariableOptions) (*WorkspaceVariable, error) {
	var (
		subject authz.Subject
		before  *WorkspaceVariable
		after   WorkspaceVariable
	)
	err := s.db.Lock(ctx, "variables", func(ctx context.Context, _ sql.Connection) (err error) {
		before, err = s.db.getWorkspaceVariable(ctx, variableID)
		if err != nil {
			return err
		}

		subject, err = s.Authorize(ctx, authz.UpdateWorkspaceVariableAction, &authz.AccessRequest{ID: &before.WorkspaceID})
		if err != nil {
			return err
		}

		workspaceVariables, err := s.db.listWorkspaceVariables(ctx, before.WorkspaceID)
		if err != nil {
			return err
		}

		// update a copy of v
		after = *before
		if err := after.update(workspaceVariables, opts); err != nil {
			return err
		}

		if err := s.db.updateVariable(ctx, after.Variable); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		s.Error(err, "updating workspace variable", "subject", subject, "variable_id", variableID)
		return nil, err
	}
	s.V(1).Info("updated workspace variable", "subject", subject, "workspace_id", after.WorkspaceID, "before", before, "after", &after)

	return &after, nil
}

func (s *Service) ListWorkspaceVariables(ctx context.Context, workspaceID resource.ID) ([]*Variable, error) {
	subject, err := s.Authorize(ctx, authz.ListWorkspaceVariablesAction, &authz.AccessRequest{ID: &workspaceID})
	if err != nil {
		return nil, err
	}

	vars, err := s.db.listWorkspaceVariables(ctx, workspaceID)
	if err != nil {
		s.Error(err, "listing workspace variables", "subject", subject, "workspace_id", workspaceID)
		return nil, err
	}

	s.V(9).Info("listed workspace variables", "subject", subject, "workspace_id", workspaceID, "count", len(vars))

	return vars, nil
}

func (s *Service) GetWorkspaceVariable(ctx context.Context, variableID resource.ID) (*WorkspaceVariable, error) {
	wv, err := s.db.getWorkspaceVariable(ctx, variableID)
	if err != nil {
		s.Error(err, "retrieving workspace variable", "variable_id", variableID)
		return nil, err
	}

	subject, err := s.Authorize(ctx, authz.ListWorkspaceVariablesAction, &authz.AccessRequest{ID: &wv.WorkspaceID})
	if err != nil {
		return nil, err
	}

	s.V(9).Info("retrieved workspace variable", "subject", subject, "workspace_id", wv.WorkspaceID, "variable", wv.Variable)

	return wv, nil
}

func (s *Service) DeleteWorkspaceVariable(ctx context.Context, variableID resource.ID) (*WorkspaceVariable, error) {
	var (
		subject authz.Subject
		wv      *WorkspaceVariable
	)
	err := s.db.Tx(ctx, func(ctx context.Context, _ sql.Connection) (err error) {
		wv, err = s.db.deleteWorkspaceVariable(ctx, variableID)
		if err != nil {
			return err
		}

		subject, err = s.Authorize(ctx, authz.DeleteWorkspaceVariableAction, &authz.AccessRequest{ID: &wv.WorkspaceID})
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		s.Error(err, "deleting workspace variable", "subject", subject, "variable_id", variableID)
		return nil, err
	}
	s.V(1).Info("deleted workspace variable", "subject", subject, "workspace_id", wv.WorkspaceID, "variable", wv.Variable)

	return wv, nil
}

func (s *Service) createVariableSet(ctx context.Context, organization string, opts CreateVariableSetOptions) (*VariableSet, error) {
	subject, err := s.Authorize(ctx, authz.CreateVariableSetAction, &authz.AccessRequest{Organization: organization})
	if err != nil {
		return nil, err
	}

	set, err := newSet(organization, opts)
	if err != nil {
		s.Error(err, "constructing variable set", "subject", subject, "organization", organization)
		return nil, err
	}

	err = s.db.Tx(ctx, func(ctx context.Context, _ sql.Connection) error {
		if err := s.db.createVariableSet(ctx, set); err != nil {
			return err
		}
		if err := s.db.createVariableSetWorkspaces(ctx, set.ID, opts.Workspaces); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		s.Error(err, "creating variable set", "subject", subject, "set", set)
		return nil, err
	}

	s.V(1).Info("created variable set", "subject", subject, "set", set)

	return set, nil
}

func (s *Service) updateVariableSet(ctx context.Context, setID resource.ID, opts UpdateVariableSetOptions) (*VariableSet, error) {
	var (
		subject authz.Subject
		before  *VariableSet
		after   VariableSet
	)
	err := s.db.Lock(ctx, "variables, variable_sets", func(ctx context.Context, _ sql.Connection) (err error) {
		before, err = s.db.getVariableSet(ctx, setID)
		if err != nil {
			return err
		}

		subject, err = s.Authorize(ctx, authz.UpdateVariableSetAction, &authz.AccessRequest{Organization: before.Organization})
		if err != nil {
			return err
		}

		organizationSets, err := s.db.listVariableSets(ctx, before.Organization)
		if err != nil {
			return err
		}

		// update copy of set
		after = *before
		if err := after.updateProperties(organizationSets, opts); err != nil {
			return err
		}

		return s.db.updateVariableSet(ctx, &after)
	})
	if err != nil {
		s.Error(err, "updating variable set", "subject", subject, "set_id", setID)
		return nil, err
	}
	s.V(1).Info("updated variable set", "subject", subject, "before", before, "after", &after)

	return &after, nil
}

func (s *Service) listVariableSets(ctx context.Context, organization string) ([]*VariableSet, error) {
	subject, err := s.Authorize(ctx, authz.ListVariableSetsAction, &authz.AccessRequest{Organization: organization})
	if err != nil {
		return nil, err
	}

	sets, err := s.db.listVariableSets(ctx, organization)
	if err != nil {
		s.Error(err, "listing variable sets", "subject", subject, "organization", organization)
		return nil, err
	}
	s.V(9).Info("listed variable sets", "subject", subject, "organization", organization, "count", len(sets))

	return sets, nil
}

func (s *Service) listWorkspaceVariableSets(ctx context.Context, workspaceID resource.ID) ([]*VariableSet, error) {
	subject, err := s.Authorize(ctx, authz.ListVariableSetsAction, &authz.AccessRequest{ID: &workspaceID})
	if err != nil {
		return nil, err
	}

	sets, err := s.db.listVariableSetsByWorkspace(ctx, workspaceID)
	if err != nil {
		s.Error(err, "listing variable sets", "subject", subject, "workspace_id", workspaceID)
		return nil, err
	}
	s.V(9).Info("listed variable sets", "subject", subject, "workspace_id", workspaceID, "count", len(sets))

	return sets, nil
}

func (s *Service) getVariableSet(ctx context.Context, setID resource.ID) (*VariableSet, error) {
	set, err := s.db.getVariableSet(ctx, setID)
	if err != nil {
		s.Error(err, "retrieving variable set", "set_id", setID)
		return nil, err
	}

	subject, err := s.Authorize(ctx, authz.GetVariableSetAction, &authz.AccessRequest{Organization: set.Organization})
	if err != nil {
		s.Error(err, "retrieving variable set", "subject", subject, "set", set)
		return nil, err
	}
	s.V(9).Info("retrieved variable set", "subject", subject, "set", set)

	return set, nil
}

func (s *Service) getVariableSetByVariableID(ctx context.Context, variableID resource.ID) (*VariableSet, error) {
	set, err := s.db.getVariableSetByVariableID(ctx, variableID)
	if err != nil {
		s.Error(err, "retrieving variable set", "variable_id", variableID)
		return nil, err
	}

	subject, err := s.Authorize(ctx, authz.GetVariableSetVariableAction, &authz.AccessRequest{Organization: set.Organization})
	if err != nil {
		return nil, err
	}

	s.V(1).Info("retrieved variable set", "subject", subject, "set", set, "variable")

	return set, nil
}

func (s *Service) deleteVariableSet(ctx context.Context, setID resource.ID) (*VariableSet, error) {
	set, err := s.db.getVariableSet(ctx, setID)
	if err != nil {
		s.Error(err, "retrieving variable set", "set_id", setID)
		return nil, err
	}

	subject, err := s.Authorize(ctx, authz.DeleteVariableSetAction, &authz.AccessRequest{Organization: set.Organization})
	if err != nil {
		return nil, err
	}

	if err := s.db.deleteVariableSet(ctx, setID); err != nil {
		s.Error(err, "deleting variable set", "subject", subject, "set", set)
		return nil, err
	}
	s.V(1).Info("deleted variable set", "subject", subject, "set", set)

	return set, nil
}

func (s *Service) createVariableSetVariable(ctx context.Context, setID resource.ID, opts CreateVariableOptions) (*Variable, error) {
	var (
		subject authz.Subject
		set     *VariableSet
		v       *Variable
	)
	err := s.db.Lock(ctx, "variables", func(ctx context.Context, _ sql.Connection) (err error) {
		set, err = s.db.getVariableSet(ctx, setID)
		if err != nil {
			return err
		}

		subject, err = s.Authorize(ctx, authz.AddVariableToSetAction, &authz.AccessRequest{Organization: set.Organization})
		if err != nil {
			return err
		}

		organizationSets, err := s.db.listVariableSets(ctx, set.Organization)
		if err != nil {
			return err
		}

		v, err = set.addVariable(organizationSets, opts)
		if err != nil {
			return err
		}

		if err := s.db.addVariableToSet(ctx, setID, v); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		s.Error(err, "adding variable to set", "set_id", setID)
		return nil, err
	}

	s.V(1).Info("added variable to set", "subject", subject, "set", set, "variable", v)

	return v, nil
}

func (s *Service) updateVariableSetVariable(ctx context.Context, variableID resource.ID, opts UpdateVariableOptions) (*VariableSet, error) {
	var (
		subject authz.Subject
		set     *VariableSet
		before  Variable
		after   *Variable
	)
	err := s.db.Lock(ctx, "variables", func(ctx context.Context, _ sql.Connection) (err error) {
		set, err = s.db.getVariableSetByVariableID(ctx, variableID)
		if err != nil {
			return err
		}
		subject, err = s.Authorize(ctx, authz.UpdateVariableSetAction, &authz.AccessRequest{Organization: set.Organization})
		if err != nil {
			return err
		}

		organizationSets, err := s.db.listVariableSets(ctx, set.Organization)
		if err != nil {
			return err
		}

		// make copy of variable before updating
		before = *set.getVariable(variableID)
		after, err = set.updateVariable(organizationSets, variableID, opts)
		if err != nil {
			return err
		}

		if err := s.db.updateVariable(ctx, after); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		s.Error(err, "updating variable set variable", "subject", subject, "variable_id", variableID)
		return nil, err
	}
	s.V(1).Info("updated variable set variable", "subject", subject, "set", set, "before", &before, "after", after)

	return set, nil
}

func (s *Service) deleteVariableSetVariable(ctx context.Context, variableID resource.ID) (*VariableSet, error) {
	set, err := s.db.getVariableSetByVariableID(ctx, variableID)
	if err != nil {
		return nil, err
	}

	subject, err := s.Authorize(ctx, authz.RemoveVariableFromSetAction, &authz.AccessRequest{Organization: set.Organization})
	if err != nil {
		return nil, err
	}

	v := set.getVariable(variableID)

	if err := s.db.deleteVariable(ctx, variableID); err != nil {
		s.Error(err, "deleting variable from set", "subject", subject, "variable", v, "set", set)
		return nil, err
	}
	s.V(1).Info("deleted variable from set", "subject", subject, "variable", v, "set", set)

	return set, nil
}

func (s *Service) applySetToWorkspaces(ctx context.Context, setID resource.ID, workspaceIDs []resource.ID) error {
	// retrieve set first in order to retrieve organization name for authorization
	set, err := s.db.getVariableSet(ctx, setID)
	if err != nil {
		return err
	}

	subject, err := s.Authorize(ctx, authz.ApplyVariableSetToWorkspacesAction, &authz.AccessRequest{Organization: set.Organization})
	if err != nil {
		return err
	}

	if err := s.db.createVariableSetWorkspaces(ctx, setID, workspaceIDs); err != nil {
		s.Error(err, "applying variable set to workspaces", "subject", subject, "set", set, "workspaces", workspaceIDs)
		return err
	}
	s.V(1).Info("applied variable set to workspaces", "subject", subject, "set", set, "workspaces", workspaceIDs)

	return nil
}

func (s *Service) deleteSetFromWorkspaces(ctx context.Context, setID resource.ID, workspaceIDs []resource.ID) error {
	// retrieve set first in order to retrieve organization name for authorization
	set, err := s.db.getVariableSet(ctx, setID)
	if err != nil {
		return err
	}

	subject, err := s.Authorize(ctx, authz.DeleteVariableSetFromWorkspacesAction, &authz.AccessRequest{Organization: set.Organization})
	if err != nil {
		return err
	}

	if err := s.db.deleteVariableSetWorkspaces(ctx, setID, workspaceIDs); err != nil {
		s.Error(err, "removing variable set from workspaces", "subject", subject, "set", set, "workspaces", workspaceIDs)
		return err
	}
	s.V(1).Info("removed variable set from workspaces", "subject", subject, "set", set, "workspaces", workspaceIDs)

	return nil
}
