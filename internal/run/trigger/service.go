package trigger

import (
	"context"
	"fmt"

	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/sql"
)

const (
	// Inbound filters triggers by the workspace that triggers the queried
	// workspace.
	Inbound Direction = "inbound"
	// Outbound filters triggers by workspaces that are triggered *by* the
	// queried workspace.
	Outbound Direction = "outbound"
)

type Direction string

type (
	// Alias service to permit embedding it with other services in a struct
	// without a name clash.
	RunTriggerService = Service

	Service struct {
		logger     logr.Logger
		authorizer *authz.Authorizer
		db         *pgdb
	}

	Options struct {
		DB         *sql.DB
		Logger     logr.Logger
		Authorizer *authz.Authorizer
	}

	// ListOptions filters run triggers by workspace and direction.
	ListOptions struct {
		// WorkspaceID is the workspace whose triggers are being listed.
		WorkspaceID resource.TfeID `schema:"workspace_id"`
		// Filters by direction: "inbound" (sourceables that trigger this
		// workspace) or "outbound" (workspaces this workspace triggers).
		Direction Direction `schema:"filter[run-trigger][type]"`
	}
)

func NewService(opts Options) *Service {
	svc := &Service{
		logger:     opts.Logger,
		authorizer: opts.Authorizer,
		db:         &pgdb{opts.DB},
	}
	// Register parent resolver so the authorizer can resolve run trigger -> workspace.
	opts.Authorizer.RegisterParentResolver(resource.RunTriggerKind,
		func(ctx context.Context, id resource.ID) (resource.ID, error) {
			t, err := svc.db.get(ctx, id)
			if err != nil {
				return nil, err
			}
			return t.WorkspaceID, nil
		},
	)
	return svc
}

func (s *Service) CreateRunTrigger(ctx context.Context, workspaceID, sourceableWorkspaceID resource.TfeID) (*Trigger, error) {
	// User must have appropriate perm on the specified workspace and permission
	// to read runs for the soureable workspace.
	subject, err := s.authorizer.Authorize(ctx, authz.CreateRunTriggerAction, workspaceID)
	if err != nil {
		return nil, err
	}
	_, err = s.authorizer.Authorize(ctx, authz.GetRunAction, sourceableWorkspaceID)
	if err != nil {
		return nil, err
	}

	t, err := newTrigger(workspaceID, sourceableWorkspaceID)
	if err != nil {
		s.logger.Error(err, "constructing run trigger", "subject", subject)
		return nil, err
	}
	if err := s.db.create(ctx, t); err != nil {
		s.logger.Error(err, "creating run trigger", "subject", subject)
		return nil, err
	}
	s.logger.V(0).Info("created run trigger", "trigger", t.ID, "workspace", workspaceID, "sourceable", sourceableWorkspaceID, "subject", subject)
	return t, nil
}

func (s *Service) ListRunTriggers(ctx context.Context, opts ListOptions) ([]*Trigger, error) {
	subject, err := s.authorizer.Authorize(ctx, authz.ListRunTriggersAction, opts.WorkspaceID)
	if err != nil {
		return nil, err
	}
	var triggers []*Trigger
	switch opts.Direction {
	case Outbound:
		triggers, err = s.db.listBySourceableWorkspaceID(ctx, opts.WorkspaceID)
	case Inbound:
		triggers, err = s.db.listByWorkspaceID(ctx, opts.WorkspaceID)
	default:
		return nil, fmt.Errorf("invalid direction: %s", opts.Direction)
	}
	if err != nil {
		s.logger.Error(err, "listing run triggers", "workspace", opts.WorkspaceID, "subject", subject)
		return nil, err
	}
	s.logger.V(9).Info("listed run triggers", "total", len(triggers), "workspace", opts.WorkspaceID, "subject", subject)
	return triggers, nil
}

func (s *Service) GetRunTrigger(ctx context.Context, triggerID resource.TfeID) (*Trigger, error) {
	subject, err := s.authorizer.Authorize(ctx, authz.GetRunTriggerAction, triggerID)
	if err != nil {
		return nil, err
	}
	t, err := s.db.get(ctx, triggerID)
	if err != nil {
		s.logger.Error(err, "retrieving run trigger", "trigger", triggerID, "subject", subject)
		return nil, err
	}
	s.logger.V(9).Info("retrieved run trigger", "trigger", t.ID, "subject", subject)
	return t, nil
}

func (s *Service) DeleteRunTrigger(ctx context.Context, triggerID resource.TfeID) error {
	subject, err := s.authorizer.Authorize(ctx, authz.DeleteRunTriggerAction, triggerID)
	if err != nil {
		return err
	}
	if err := s.db.delete(ctx, triggerID); err != nil {
		s.logger.Error(err, "deleting run trigger", "trigger", triggerID, "subject", subject)
		return err
	}
	s.logger.V(0).Info("deleted run trigger", "trigger", triggerID, "subject", subject)
	return nil
}
