package variable

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/http/html"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/rbac"
	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/tfeapi"
	"github.com/leg100/otf/internal/workspace"
	"github.com/pkg/errors"
)

type (
	VariableService = Service

	Service interface {
		CreateVariable(ctx context.Context, workspaceID string, opts CreateVariableOptions) (*Variable, error)
		ListVariables(ctx context.Context, workspaceID string) ([]*Variable, error)
		GetVariable(ctx context.Context, variableID string) (*Variable, error)
		UpdateVariable(ctx context.Context, variableID string, opts UpdateVariableOptions) (*Variable, error)
		DeleteVariable(ctx context.Context, variableID string) (*Variable, error)

		CreateVariableSet(ctx context.Context, organization string, opts CreateVariableSetOptions) (*VariableSet, error)
		ListVariableSets(ctx context.Context, organization string) ([]*VariableSet, error)
		UpdateVariableSet(ctx context.Context, setID string, opts UpdateVariableSetOptions) (*VariableSet, error)
		GetVariableSet(ctx context.Context, setID string) (*VariableSet, error)
		DeleteVariableSet(ctx context.Context, setID string) error
	}

	service struct {
		logr.Logger

		db           *pgdb
		web          *web
		api          *tfe
		workspace    internal.Authorizer
		organization internal.Authorizer

		*factory
	}

	Options struct {
		WorkspaceAuthorizer internal.Authorizer
		WorkspaceService    workspace.Service

		*sql.DB
		*tfeapi.Responder
		html.Renderer
		logr.Logger
	}
)

func NewService(opts Options) *service {
	svc := service{
		Logger:       opts.Logger,
		db:           &pgdb{opts.DB},
		factory:      &factory{generateVersion: versionGenerator},
		workspace:    opts.WorkspaceAuthorizer,
		organization: &organization.Authorizer{Logger: opts.Logger},
	}

	svc.web = &web{
		Renderer: opts.Renderer,
		Service:  opts.WorkspaceService,
		svc:      &svc,
	}
	svc.api = &tfe{
		Service:   &svc,
		Responder: opts.Responder,
	}

	return &svc
}

func (s *service) AddHandlers(r *mux.Router) {
	s.web.addHandlers(r)
	s.api.addHandlers(r)
}

func (s *service) CreateVariable(ctx context.Context, workspaceID string, opts CreateVariableOptions) (*Variable, error) {
	subject, err := s.workspace.CanAccess(ctx, rbac.CreateVariableAction, workspaceID)
	if err != nil {
		return nil, err
	}

	v, err := s.new(workspaceID, opts)
	if err != nil {
		s.Error(err, "constructing variable", "subject", subject, "workspace", workspaceID, "key", opts.Key)
		return nil, err
	}

	if err := s.db.create(ctx, v); err != nil {
		s.Error(err, "creating variable", "subject", subject, "variable", v)
		return nil, err
	}

	s.V(1).Info("created variable", "subject", subject, "variable", v)

	return v, nil
}

func (s *service) ListVariables(ctx context.Context, workspaceID string) ([]*Variable, error) {
	subject, err := s.workspace.CanAccess(ctx, rbac.ListVariablesAction, workspaceID)
	if err != nil {
		return nil, err
	}

	variables, err := s.db.list(ctx, workspaceID)
	if err != nil {
		s.Error(err, "listing variables", "subject", subject, "workspace_id", workspaceID)
		return nil, err
	}

	s.V(9).Info("listed variables", "subject", subject, "workspace_id", workspaceID)

	return variables, nil
}

func (s *service) GetVariable(ctx context.Context, variableID string) (*Variable, error) {
	// retrieve variable first in order to retrieve workspace ID for authorization
	variable, err := s.db.get(ctx, variableID)
	if err != nil {
		s.Error(err, "retrieving variable", "workspace_id", variableID)
		return nil, err
	}

	subject, err := s.workspace.CanAccess(ctx, rbac.GetVariableAction, variable.WorkspaceID)
	if err != nil {
		return nil, err
	}

	s.V(9).Info("retrieved variable", "subject", subject, "variable", variable)

	return variable, nil
}

func (s *service) UpdateVariable(ctx context.Context, variableID string, opts UpdateVariableOptions) (*Variable, error) {
	// retrieve existing in order to retrieve workspace ID for authorization
	existing, err := s.db.get(ctx, variableID)
	if err != nil {
		return nil, errors.Wrap(err, "retrieving variable")
	}

	subject, err := s.workspace.CanAccess(ctx, rbac.UpdateVariableAction, existing.WorkspaceID)
	if err != nil {
		return nil, err
	}

	updated, err := s.db.update(ctx, variableID, func(v *Variable) error {
		return s.update(v, opts)
	})
	if err != nil {
		s.Error(err, "updating variable", "subject", subject, "variable_id", variableID, "workspace_id", existing.WorkspaceID)
		return nil, err
	}
	s.V(1).Info("updated variable", "subject", subject, "before", existing, "after", updated)

	return updated, nil
}

func (s *service) DeleteVariable(ctx context.Context, variableID string) (*Variable, error) {
	// retrieve existing in order to retrieve workspace ID for authorization
	existing, err := s.db.get(ctx, variableID)
	if err != nil {
		return nil, err
	}

	subject, err := s.workspace.CanAccess(ctx, rbac.DeleteVariableAction, existing.WorkspaceID)
	if err != nil {
		return nil, err
	}

	deleted, err := s.db.delete(ctx, variableID)
	if err != nil {
		s.Error(err, "deleting variable", "subject", subject, "variable", existing)
		return nil, err
	}
	s.V(1).Info("deleted variable", "subject", subject, "variable", deleted)

	return deleted, nil
}

func (s *service) CreateVariableSet(ctx context.Context, organization string, opts CreateVariableSetOptions) (*VariableSet, error) {
	subject, err := s.organization.CanAccess(ctx, rbac.CreateVariableSetAction, organization)
	if err != nil {
		return nil, err
	}

	v, err := s.newSet(organization, opts)
	if err != nil {
		s.Error(err, "constructing variable set", "subject", subject, "organization", organization)
		return nil, err
	}

	if err := s.db.create(ctx, v); err != nil {
		s.Error(err, "creating variable set", "subject", subject, "organization", organization)
		return nil, err
	}

	s.V(1).Info("created variable set", "subject", subject, "organization", organization)

	return v, nil
}

func (s *service) DeleteVariableSet(ctx context.Context, setID string) error {
	// retrieve existing set in order to retrieve organization for authorization
	existing, err := s.db.get(ctx, setID)
	if err != nil {
		return err
	}

	subject, err := s.organization.CanAccess(ctx, rbac.DeleteVariableSetAction, existing.Organization)
	if err != nil {
		return err
	}

	deleted, err := s.db.delete(ctx, setID)
	if err != nil {
		s.Error(err, "deleting variable", "subject", subject, "variable", existing)
		return err
	}
	s.V(1).Info("deleted variable", "subject", subject, "variable", deleted)

	return nil
}
