package variable

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/http/html"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/rbac"
	"github.com/leg100/otf/internal/run"
	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/sql/sqlc"
	"github.com/leg100/otf/internal/tfeapi"
	"github.com/leg100/otf/internal/workspace"
)

type (
	Service struct {
		logr.Logger

		db           *pgdb
		web          *web
		tfeapi       *tfe
		api          *api
		workspace    internal.Authorizer
		organization internal.Authorizer
		runs         runClient
	}

	Options struct {
		WorkspaceAuthorizer internal.Authorizer
		WorkspaceService    *workspace.Service
		RunClient           runClient

		*sql.DB
		*tfeapi.Responder
		html.Renderer
		logr.Logger
	}

	runClient interface {
		Get(ctx context.Context, runID string) (*run.Run, error)
	}
)

func NewService(opts Options) *Service {
	svc := Service{
		Logger:       opts.Logger,
		db:           &pgdb{opts.DB},
		workspace:    opts.WorkspaceAuthorizer,
		organization: &organization.Authorizer{Logger: opts.Logger},
		runs:         opts.RunClient,
	}

	svc.web = &web{
		Renderer:   opts.Renderer,
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

func (s *Service) ListEffectiveVariables(ctx context.Context, runID string) ([]*Variable, error) {
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

func (s *Service) CreateWorkspaceVariable(ctx context.Context, workspaceID string, opts CreateVariableOptions) (*Variable, error) {
	subject, err := s.workspace.CanAccess(ctx, rbac.CreateWorkspaceVariableAction, workspaceID)
	if err != nil {
		return nil, err
	}

	var v *Variable
	err = s.db.Lock(ctx, "variables", func(ctx context.Context, q *sqlc.Queries) (err error) {
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

func (s *Service) UpdateWorkspaceVariable(ctx context.Context, variableID string, opts UpdateVariableOptions) (*WorkspaceVariable, error) {
	var (
		subject internal.Subject
		before  *WorkspaceVariable
		after   WorkspaceVariable
	)
	err := s.db.Lock(ctx, "variables", func(ctx context.Context, q *sqlc.Queries) (err error) {
		before, err = s.db.getWorkspaceVariable(ctx, variableID)
		if err != nil {
			return err
		}

		subject, err = s.workspace.CanAccess(ctx, rbac.UpdateWorkspaceVariableAction, before.WorkspaceID)
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

func (s *Service) ListWorkspaceVariables(ctx context.Context, workspaceID string) ([]*Variable, error) {
	subject, err := s.workspace.CanAccess(ctx, rbac.ListWorkspaceVariablesAction, workspaceID)
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

func (s *Service) GetWorkspaceVariable(ctx context.Context, variableID string) (*WorkspaceVariable, error) {
	wv, err := s.db.getWorkspaceVariable(ctx, variableID)
	if err != nil {
		s.Error(err, "retrieving workspace variable", "variable_id", variableID)
		return nil, err
	}

	subject, err := s.workspace.CanAccess(ctx, rbac.ListWorkspaceVariablesAction, wv.WorkspaceID)
	if err != nil {
		return nil, err
	}

	s.V(9).Info("retrieved workspace variable", "subject", subject, "workspace_id", wv.WorkspaceID, "variable", wv.Variable)

	return wv, nil
}

func (s *Service) DeleteWorkspaceVariable(ctx context.Context, variableID string) (*WorkspaceVariable, error) {
	var (
		subject internal.Subject
		wv      *WorkspaceVariable
	)
	err := s.db.Tx(ctx, func(ctx context.Context, q *sqlc.Queries) (err error) {
		wv, err = s.db.deleteWorkspaceVariable(ctx, variableID)
		if err != nil {
			return err
		}

		subject, err = s.workspace.CanAccess(ctx, rbac.DeleteWorkspaceVariableAction, wv.WorkspaceID)
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
	subject, err := s.organization.CanAccess(ctx, rbac.CreateVariableSetAction, organization)
	if err != nil {
		return nil, err
	}

	set, err := newSet(organization, opts)
	if err != nil {
		s.Error(err, "constructing variable set", "subject", subject, "organization", organization)
		return nil, err
	}

	err = s.db.Tx(ctx, func(ctx context.Context, q *sqlc.Queries) error {
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

func (s *Service) updateVariableSet(ctx context.Context, setID string, opts UpdateVariableSetOptions) (*VariableSet, error) {
	var (
		subject internal.Subject
		before  *VariableSet
		after   VariableSet
	)
	err := s.db.Lock(ctx, "variables, variable_sets", func(ctx context.Context, q *sqlc.Queries) (err error) {
		before, err = s.db.getVariableSet(ctx, setID)
		if err != nil {
			return err
		}

		subject, err = s.organization.CanAccess(ctx, rbac.UpdateVariableSetAction, before.Organization)
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
	subject, err := s.organization.CanAccess(ctx, rbac.ListVariableSetsAction, organization)
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

func (s *Service) listWorkspaceVariableSets(ctx context.Context, workspaceID string) ([]*VariableSet, error) {
	subject, err := s.workspace.CanAccess(ctx, rbac.ListVariableSetsAction, workspaceID)
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

func (s *Service) getVariableSet(ctx context.Context, setID string) (*VariableSet, error) {
	set, err := s.db.getVariableSet(ctx, setID)
	if err != nil {
		s.Error(err, "retrieving variable set", "set_id", setID)
		return nil, err
	}

	subject, err := s.organization.CanAccess(ctx, rbac.GetVariableSetAction, set.Organization)
	if err != nil {
		s.Error(err, "retrieving variable set", "subject", subject, "set", set)
		return nil, err
	}
	s.V(9).Info("retrieved variable set", "subject", subject, "set", set)

	return set, nil
}

func (s *Service) getVariableSetByVariableID(ctx context.Context, variableID string) (*VariableSet, error) {
	set, err := s.db.getVariableSetByVariableID(ctx, variableID)
	if err != nil {
		s.Error(err, "retrieving variable set", "variable_id", variableID)
		return nil, err
	}

	subject, err := s.organization.CanAccess(ctx, rbac.GetVariableSetVariableAction, set.Organization)
	if err != nil {
		return nil, err
	}

	s.V(1).Info("retrieved variable set", "subject", subject, "set", set, "variable")

	return set, nil
}

func (s *Service) deleteVariableSet(ctx context.Context, setID string) (*VariableSet, error) {
	set, err := s.db.getVariableSet(ctx, setID)
	if err != nil {
		s.Error(err, "retrieving variable set", "set_id", setID)
		return nil, err
	}

	subject, err := s.organization.CanAccess(ctx, rbac.DeleteVariableSetAction, set.Organization)
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

func (s *Service) createVariableSetVariable(ctx context.Context, setID string, opts CreateVariableOptions) (*Variable, error) {
	var (
		subject internal.Subject
		set     *VariableSet
		v       *Variable
	)
	err := s.db.Lock(ctx, "variables", func(ctx context.Context, q *sqlc.Queries) (err error) {
		set, err = s.db.getVariableSet(ctx, setID)
		if err != nil {
			return err
		}

		subject, err = s.organization.CanAccess(ctx, rbac.AddVariableToSetAction, set.Organization)
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

func (s *Service) updateVariableSetVariable(ctx context.Context, variableID string, opts UpdateVariableOptions) (*VariableSet, error) {
	var (
		subject internal.Subject
		set     *VariableSet
		before  Variable
		after   *Variable
	)
	err := s.db.Lock(ctx, "variables", func(ctx context.Context, q *sqlc.Queries) (err error) {
		set, err = s.db.getVariableSetByVariableID(ctx, variableID)
		if err != nil {
			return err
		}
		subject, err = s.organization.CanAccess(ctx, rbac.UpdateVariableSetAction, set.Organization)
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

func (s *Service) deleteVariableSetVariable(ctx context.Context, variableID string) (*VariableSet, error) {
	set, err := s.db.getVariableSetByVariableID(ctx, variableID)
	if err != nil {
		return nil, err
	}

	subject, err := s.organization.CanAccess(ctx, rbac.RemoveVariableFromSetAction, set.Organization)
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

func (s *Service) applySetToWorkspaces(ctx context.Context, setID string, workspaceIDs []string) error {
	// retrieve set first in order to retrieve organization name for authorization
	set, err := s.db.getVariableSet(ctx, setID)
	if err != nil {
		return err
	}

	subject, err := s.organization.CanAccess(ctx, rbac.ApplyVariableSetToWorkspacesAction, set.Organization)
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

func (s *Service) deleteSetFromWorkspaces(ctx context.Context, setID string, workspaceIDs []string) error {
	// retrieve set first in order to retrieve organization name for authorization
	set, err := s.db.getVariableSet(ctx, setID)
	if err != nil {
		return err
	}

	subject, err := s.organization.CanAccess(ctx, rbac.DeleteVariableSetFromWorkspacesAction, set.Organization)
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
