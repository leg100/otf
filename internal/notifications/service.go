package notifications

import (
	"context"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/pubsub"
	"github.com/leg100/otf/internal/rbac"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/tfeapi"
)

type (
	Service struct {
		logr.Logger

		workspaceAuthorizer internal.Authorizer // authorize workspaces actions
		db                  *pgdb
		api                 *tfe
		broker              *pubsub.Broker[*Config]
	}

	Options struct {
		*sql.DB
		*sql.Listener
		*tfeapi.Responder
		logr.Logger

		WorkspaceAuthorizer internal.Authorizer
	}
)

func NewService(opts Options) *Service {
	svc := Service{
		Logger:              opts.Logger,
		workspaceAuthorizer: opts.WorkspaceAuthorizer,
		db:                  &pgdb{opts.DB},
	}
	svc.api = &tfe{
		Service:   &svc,
		Responder: opts.Responder,
	}
	// Register with broker so that it can relay run events
	svc.broker = pubsub.NewBroker(
		opts.Logger,
		opts.Listener,
		"notification_configurations",
		func(ctx context.Context, id resource.ID, action sql.Action) (*Config, error) {
			if action == sql.DeleteAction {
				return &Config{ID: id}, nil
			}
			return svc.db.get(ctx, id)
		},
	)
	return &svc
}

func (s *Service) AddHandlers(r *mux.Router) {
	s.api.addHandlers(r)
}

func (s *Service) Watch(ctx context.Context) (<-chan pubsub.Event[*Config], func()) {
	return s.broker.Subscribe(ctx)
}

func (s *Service) Create(ctx context.Context, workspaceID resource.ID, opts CreateConfigOptions) (*Config, error) {
	subject, err := s.workspaceAuthorizer.CanAccess(ctx, rbac.CreateNotificationConfigurationAction, workspaceID)
	if err != nil {
		return nil, err
	}
	nc, err := NewConfig(workspaceID, opts)
	if err != nil {
		s.Error(err, "constructing notification config", "subject", subject)
		return nil, err
	}
	if err := s.db.create(ctx, nc); err != nil {
		s.Error(err, "creating notification config", "config", nc, "subject", subject)
		return nil, err
	}
	s.Info("creating notification config", "config", nc, "subject", subject)
	return nc, nil
}

func (s *Service) Update(ctx context.Context, id resource.ID, opts UpdateConfigOptions) (*Config, error) {
	var subject internal.Subject
	updated, err := s.db.update(ctx, id, func(nc *Config) (err error) {
		subject, err = s.workspaceAuthorizer.CanAccess(ctx, rbac.UpdateNotificationConfigurationAction, nc.WorkspaceID)
		if err != nil {
			return err
		}
		return nc.update(opts)
	})
	if err != nil {
		s.Error(err, "updating notification config", "id", id, "subject", subject)
		return nil, err
	}
	s.Info("updated notification config", "updated", updated, "subject", subject)
	return updated, nil
}

func (s *Service) Get(ctx context.Context, id resource.ID) (*Config, error) {
	nc, err := s.db.get(ctx, id)
	if err != nil {
		s.Error(err, "retrieving notification config", "id", id)
		return nil, err
	}
	subject, err := s.workspaceAuthorizer.CanAccess(ctx, rbac.GetNotificationConfigurationAction, nc.WorkspaceID)
	if err != nil {
		return nil, err
	}
	s.V(9).Info("retrieved notification config", "config", nc, "subject", subject)
	return nc, nil
}

func (s *Service) List(ctx context.Context, workspaceID resource.ID) ([]*Config, error) {
	subject, err := s.workspaceAuthorizer.CanAccess(ctx, rbac.ListNotificationConfigurationsAction, workspaceID)
	if err != nil {
		return nil, err
	}
	configs, err := s.db.list(ctx, workspaceID)
	if err != nil {
		s.Error(err, "listing notification configs", "id", workspaceID)
		return nil, err
	}
	s.V(9).Info("listed notification configs", "total", len(configs), "subject", subject)
	return configs, nil
}

func (s *Service) Delete(ctx context.Context, id resource.ID) error {
	nc, err := s.db.get(ctx, id)
	if err != nil {
		s.Error(err, "retrieving notification config", "id", id)
		return err
	}
	subject, err := s.workspaceAuthorizer.CanAccess(ctx, rbac.DeleteNotificationConfigurationAction, nc.WorkspaceID)
	if err != nil {
		return err
	}
	if err := s.db.delete(ctx, id); err != nil {
		s.Error(err, "deleting notification config", "id", id)
		return err
	}
	s.Info("deleted notification config", "config", nc, "subject", subject)
	return nil
}
