package run

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"github.com/leg100/otf"
)

type Service struct {
	app

	api *api
	web *web
}

func NewService(ctx context.Context, opts Options) *Service {
	app := &Application{
		Authorizer:    opts.Authorizer,
		PubSubService: opts.PubSubService,
		Logger:        opts.Logger,
		cache:         opts.Cache,
		db:            newDB(opts.DB),
		factory: &factory{
			opts.ConfigurationVersionService,
			opts.WorkspaceService,
		},
	}
	api := &api{
		app: app,
	}
	web := &web{
		Renderer: opts.Renderer,
		app:      app,
	}
	return &Service{
		app: app,
		api: api,
		web: web,
	}
}

type Options struct {
	otf.Authorizer
	otf.Cache
	otf.DB
	otf.Renderer
	otf.PubSubService
	otf.HostnameService
	otf.WorkspaceService
	otf.ConfigurationVersionService
	logr.Logger
}

func (s *Service) AddHandlers(r *mux.Router) {
	s.api.addHandlers(r)
	s.web.addHandlers(r)
}

func (s *Service) Create(ctx context.Context, workspaceID string, opts RunCreateOptions) (*Run, error) {
	return s.create(ctx, workspaceID, opts)
}
