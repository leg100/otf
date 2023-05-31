package notifications

import (
	"context"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/pubsub"
	"github.com/leg100/otf/internal/rbac"
	"github.com/leg100/otf/internal/workspace"
)

type (
	NotificationService = Service

	Service interface {
		CreateNotificationConfiguration(ctx context.Context, workspaceID string, opts CreateConfigOptions) (*Config, error)
		UpdateNotificationConfiguration(ctx context.Context, id string, opts UpdateConfigOptions) (*Config, error)
		GetNotificationConfiguration(ctx context.Context, id string) (*Config, error)
		ListNotificationConfigurations(ctx context.Context, workspaceID string) ([]*Config, error)
		DeleteNotificationConfiguration(ctx context.Context, id string) error
	}

	service struct {
		logr.Logger
		pubsub.PubSubService
		workspace.WorkspaceService
		internal.HostnameService // for including a link in the notification

		workspace internal.Authorizer // authorize workspaces actions
		db        *pgdb
	}

	Options struct {
		internal.DB
		*pubsub.Broker
		logr.Logger
		WorkspaceAuthorizer internal.Authorizer
		workspace.WorkspaceService
		internal.HostnameService // for including a link in the notification
	}
)

func NewService(opts Options) *service {
	svc := service{
		Logger:           opts.Logger,
		PubSubService:    opts.Broker,
		workspace:        opts.WorkspaceAuthorizer,
		db:               &pgdb{opts.DB},
		HostnameService:  opts.HostnameService,
		WorkspaceService: opts.WorkspaceService,
	}
	// Register with broker so that it can relay events
	opts.Register("notification_configurations", svc.db)
	return &svc
}

func (s *service) CreateNotificationConfiguration(ctx context.Context, workspaceID string, opts CreateConfigOptions) (*Config, error) {
	subject, err := s.workspace.CanAccess(ctx, rbac.CreateNotificationConfigurationAction, workspaceID)
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
	s.Publish(pubsub.NewCreatedEvent(nc))
	return nc, nil
}

func (s *service) UpdateNotificationConfiguration(ctx context.Context, id string, opts UpdateConfigOptions) (*Config, error) {
	var subject internal.Subject
	updated, err := s.db.update(ctx, id, func(nc *Config) (err error) {
		subject, err = s.workspace.CanAccess(ctx, rbac.UpdateNotificationConfigurationAction, nc.WorkspaceID)
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

func (s *service) GetNotificationConfiguration(ctx context.Context, id string) (*Config, error) {
	nc, err := s.db.get(ctx, id)
	if err != nil {
		s.Error(err, "retrieving notification config", "id", id)
		return nil, err
	}
	subject, err := s.workspace.CanAccess(ctx, rbac.GetNotificationConfigurationAction, nc.WorkspaceID)
	if err != nil {
		return nil, err
	}
	s.V(9).Info("retrieved notification config", "config", nc, "subject", subject)
	return nc, nil
}

func (s *service) ListNotificationConfigurations(ctx context.Context, workspaceID string) ([]*Config, error) {
	subject, err := s.workspace.CanAccess(ctx, rbac.ListNotificationConfigurationsAction, workspaceID)
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

func (s *service) DeleteNotificationConfiguration(ctx context.Context, id string) error {
	nc, err := s.db.get(ctx, id)
	if err != nil {
		s.Error(err, "retrieving notification config", "id", id)
		return err
	}
	subject, err := s.workspace.CanAccess(ctx, rbac.DeleteNotificationConfigurationAction, nc.WorkspaceID)
	if err != nil {
		return err
	}
	if err := s.db.delete(ctx, id); err != nil {
		s.Error(err, "deleting notification config", "id", id)
		return err
	}
	s.Info("deleted notification config", "config", nc, "subject", subject)
	s.Publish(pubsub.NewDeletedEvent(nc))
	return nil
}
