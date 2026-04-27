package variable

import (
	"context"
	"fmt"

	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/run"
	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/workspace"
)

type (
	// Alias service to permit embedding it with other services in a struct
	// without a name clash.
	VariableService = Service

	Service struct {
		logr.Logger
		*authz.Authorizer

		db   *pgdb
		runs runClient
	}

	Options struct {
		WorkspaceService *workspace.Service
		RunClient        runClient
		Authorizer       *authz.Authorizer
		DB               *sql.DB
		Logger           logr.Logger
	}

	runClient interface {
		GetRun(ctx context.Context, runID resource.TfeID) (*run.Run, error)
	}
)

func NewService(opts Options) *Service {
	db := &pgdb{opts.DB}
	svc := Service{
		Logger:     opts.Logger,
		Authorizer: opts.Authorizer,
		db:         db,
		runs:       opts.RunClient,
	}

	// Provide a means of looking up a variables's parent resource ID.
	opts.Authorizer.RegisterParentResolver(resource.VariableKind,
		func(ctx context.Context, variableID resource.ID) (resource.ID, error) {
			// NOTE: we look up directly in the database rather than via
			// service call to avoid a recursion loop.
			v, err := db.getVariable(ctx, variableID)
			if err != nil {
				return nil, err
			}
			return v.ParentID, nil
		},
	)

	return &svc
}

func (s *Service) ListEffectiveVariables(ctx context.Context, runID resource.TfeID) ([]*Variable, error) {
	run, err := s.runs.GetRun(ctx, runID)
	if err != nil {
		return nil, err
	}
	sets, err := s.ListWorkspaceVariableSets(ctx, run.WorkspaceID)
	if err != nil {
		return nil, err
	}
	vars, err := s.ListWorkspaceVariables(ctx, run.WorkspaceID)
	if err != nil {
		return nil, err
	}
	return Merge(sets, vars, run), nil
}

func (s *Service) CreateVariable(ctx context.Context, parentID resource.TfeID, opts CreateVariableOptions) (*Variable, error) {
	// TODO: introduce CreateVariableAction
	subject, err := s.Authorize(ctx, authz.CreateWorkspaceVariableAction, parentID)
	if err != nil {
		return nil, err
	}

	var v *Variable
	err = s.db.Lock(ctx, "variables", func(ctx context.Context) (err error) {
		switch parentID.Kind() {
		case resource.WorkspaceKind:
			workspaceVariables, err := s.db.listWorkspaceVariables(ctx, parentID)
			if err != nil {
				return err
			}
			v, err = newVariable(workspaceVariables, opts)
			if err != nil {
				return err
			}
			err = s.db.createWorkspaceVariable(ctx, parentID, v)
			if err != nil {
				return err
			}
		case resource.VariableSetKind:
			set, err := s.db.getVariableSet(ctx, parentID)
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

			if err := s.db.addVariableToSet(ctx, set.ID, v); err != nil {
				return err
			}
		default:
			return fmt.Errorf("invalid variable parent kind: %s", parentID.Kind())
		}
		return nil
	})
	if err != nil {
		s.Error(err, "creating variable", "subject", subject, "variable", v)
		return nil, err
	}

	s.V(1).Info("created variable", "subject", subject, "variable", v)

	return v, nil
}

func (s *Service) UpdateVariable(ctx context.Context, variableID resource.TfeID, opts UpdateVariableOptions) (*Variable, error) {
	// TODO: introduce UpdateVariableAction
	subject, err := s.Authorize(ctx, authz.UpdateWorkspaceVariableAction, variableID)
	if err != nil {
		return nil, err
	}

	var (
		before *Variable
		after  *Variable
	)
	// Lock variables table because the update first needs to check there are no
	// conflicts with other variables in the table.
	err = s.db.Lock(ctx, "variables", func(ctx context.Context) (err error) {
		before, err = s.db.getVariable(ctx, variableID)
		if err != nil {
			return err
		}
		// update a copy of v
		after = new(*before)

		switch before.ParentID.Kind() {
		case resource.WorkspaceKind:
			workspaceVariables, err := s.db.listWorkspaceVariables(ctx, before.ParentID)
			if err != nil {
				return err
			}
			if err := after.update(workspaceVariables, opts); err != nil {
				return err
			}
		case resource.VariableSetKind:
			set, err := s.db.getVariableSetByVariableID(ctx, variableID)
			if err != nil {
				return err
			}
			organizationSets, err := s.db.listVariableSets(ctx, set.Organization)
			if err != nil {
				return err
			}
			after, err = set.updateVariable(organizationSets, variableID, opts)
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf("invalid variable parent kind: %s", before.ParentID.Kind())
		}

		if err := s.db.updateVariable(ctx, after); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		s.Error(err, "updating variable", "subject", subject, "variable_id", variableID)
		return nil, err
	}
	s.V(1).Info("updated variable", "subject", subject, "before", before, "after", after)

	return after, nil
}

func (s *Service) ListWorkspaceVariables(ctx context.Context, workspaceID resource.TfeID) ([]*Variable, error) {
	subject, err := s.Authorize(ctx, authz.ListWorkspaceVariablesAction, workspaceID)
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

func (s *Service) GetVariable(ctx context.Context, variableID resource.TfeID) (*Variable, error) {
	subject, err := s.Authorize(ctx, authz.GetWorkspaceVariableAction, variableID)
	if err != nil {
		return nil, err
	}

	v, err := s.db.getVariable(ctx, variableID)
	if err != nil {
		s.Error(err, "retrieving variable", "variable_id", variableID)
		return nil, err
	}

	s.V(9).Info("retrieved variable", "subject", subject, "variable", v)

	return v, nil
}

func (s *Service) DeleteVariable(ctx context.Context, variableID resource.TfeID) (*Variable, error) {
	// TODO: replace with DeleteVariableAction
	subject, err := s.Authorize(ctx, authz.DeleteWorkspaceVariableAction, variableID)
	if err != nil {
		return nil, err
	}
	v, err := s.db.deleteVariable(ctx, variableID)
	if err != nil {
		s.Error(err, "deleting variable", "subject", subject, "variable_id", variableID)
		return nil, err
	}
	s.V(1).Info("deleted variable", "subject", subject, "variable", v)

	return v, nil
}

func (s *Service) CreateVariableSet(ctx context.Context, organization organization.Name, opts CreateVariableSetOptions) (*VariableSet, error) {
	subject, err := s.Authorize(ctx, authz.CreateVariableSetAction, organization)
	if err != nil {
		return nil, err
	}

	set, err := newSet(organization, opts)
	if err != nil {
		s.Error(err, "constructing variable set", "subject", subject, "organization", organization)
		return nil, err
	}

	err = s.db.Tx(ctx, func(ctx context.Context) error {
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

func (s *Service) UpdateVariableSet(ctx context.Context, setID resource.TfeID, opts UpdateVariableSetOptions) (*VariableSet, error) {
	var (
		subject authz.Subject
		before  *VariableSet
		after   VariableSet
	)
	err := s.db.Lock(ctx, "variables, variable_sets", func(ctx context.Context) (err error) {
		before, err = s.db.getVariableSet(ctx, setID)
		if err != nil {
			return fmt.Errorf("retrieving variable set: %w", err)
		}

		subject, err = s.Authorize(ctx, authz.UpdateVariableSetAction, &before.Organization)
		if err != nil {
			return err
		}

		organizationSets, err := s.db.listVariableSets(ctx, before.Organization)
		if err != nil {
			return fmt.Errorf("listing variable sets: %w", err)
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

func (s *Service) ListVariableSets(ctx context.Context, organization organization.Name) ([]*VariableSet, error) {
	subject, err := s.Authorize(ctx, authz.ListVariableSetsAction, organization)
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

func (s *Service) ListWorkspaceVariableSets(ctx context.Context, workspaceID resource.TfeID) ([]*VariableSet, error) {
	subject, err := s.Authorize(ctx, authz.ListVariableSetsAction, workspaceID)
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

func (s *Service) GetVariableSet(ctx context.Context, setID resource.TfeID) (*VariableSet, error) {
	set, err := s.db.getVariableSet(ctx, setID)
	if err != nil {
		s.Error(err, "retrieving variable set", "set_id", setID)
		return nil, err
	}

	subject, err := s.Authorize(ctx, authz.GetVariableSetAction, &set.Organization)
	if err != nil {
		s.Error(err, "retrieving variable set", "subject", subject, "set", set)
		return nil, err
	}
	s.V(9).Info("retrieved variable set", "subject", subject, "set", set)

	return set, nil
}

func (s *Service) GetVariableSetByVariableID(ctx context.Context, variableID resource.TfeID) (*VariableSet, error) {
	set, err := s.db.getVariableSetByVariableID(ctx, variableID)
	if err != nil {
		s.Error(err, "retrieving variable set", "variable_id", variableID)
		return nil, err
	}

	subject, err := s.Authorize(ctx, authz.GetVariableSetVariableAction, &set.Organization)
	if err != nil {
		return nil, err
	}

	s.V(1).Info("retrieved variable set", "subject", subject, "set", set, "variable")

	return set, nil
}

func (s *Service) DeleteVariableSet(ctx context.Context, setID resource.TfeID) (*VariableSet, error) {
	set, err := s.db.getVariableSet(ctx, setID)
	if err != nil {
		s.Error(err, "retrieving variable set", "set_id", setID)
		return nil, err
	}

	subject, err := s.Authorize(ctx, authz.DeleteVariableSetAction, &set.Organization)
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

func (s *Service) ApplySetToWorkspaces(ctx context.Context, setID resource.TfeID, workspaceIDs []resource.TfeID) error {
	// retrieve set first in order to retrieve organization name for authorization
	set, err := s.db.getVariableSet(ctx, setID)
	if err != nil {
		return err
	}

	subject, err := s.Authorize(ctx, authz.ApplyVariableSetToWorkspacesAction, &set.Organization)
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

func (s *Service) DeleteSetFromWorkspaces(ctx context.Context, setID resource.TfeID, workspaceIDs []resource.TfeID) error {
	// retrieve set first in order to retrieve organization name for authorization
	set, err := s.db.getVariableSet(ctx, setID)
	if err != nil {
		return err
	}

	subject, err := s.Authorize(ctx, authz.DeleteVariableSetFromWorkspacesAction, &set.Organization)
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
