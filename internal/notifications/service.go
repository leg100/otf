package notifications

import (
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/logr"
)

type (
	NotificationService = Service

	Service interface{}

	service struct {
		logr.Logger
		internal.Subscriber

		workspace internal.Authorizer // authorize workspaces actions
		db        *pgdb
	}

	Options struct {
		internal.DB
		internal.Subscriber
		logr.Logger
		WorkspaceAuthorizer internal.Authorizer
	}
)

func NewService(opts Options) *service {
	svc := service{
		Logger:     opts.Logger,
		Subscriber: opts.Subscriber,
		workspace:  opts.WorkspaceAuthorizer,
		db:         &pgdb{opts.DB},
	}
	return &svc
}
