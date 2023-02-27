package variable

import (
	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"github.com/leg100/otf"
)

type Service struct {
	api *api
	web *web
}

func NewService(opts Options) *Service {
	app := &app{
		WorkspaceAuthorizer: opts.WorkspaceAuthorizer,
		db:                  newPGDB(opts.Database),
		Logger:              opts.Logger,
	}

	return &Service{
		api: &api{
			app: app,
		},
		web: &web{
			app:              app,
			Renderer:         opts.Renderer,
			WorkspaceService: opts.WorkspaceService,
		},
	}
}

type Options struct {
	otf.WorkspaceAuthorizer
	otf.Database
	otf.Renderer
	otf.WorkspaceService
	logr.Logger
}

func (s *Service) AddHandlers(r *mux.Router) {
	s.api.addHandlers(r)
	s.web.addHandlers(r)
}
