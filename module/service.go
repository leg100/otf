package module

import (
	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/leg100/otf/cloud"
	"github.com/leg100/surl"
)

type Service struct {
	*app

	api *api
	web *app
}

func NewService(opts Options) *Service {
	app := &app{
		OrganizationAuthorizer: opts.OrganizationAuthorizer,
		pgdb:                   &pgdb{opts.DB},
		Logger:                 opts.Logger,
	}

	return &Service{
		app: app,
		web: &web{
			Renderer: opts.Renderer,
			app:      app,
		},
	}
}

type Options struct {
	otf.OrganizationAuthorizer
	cloud.Service
	otf.DB
	*surl.Signer
	otf.Renderer
	logr.Logger
}

func (s *Service) AddHandlers(r *mux.Router) {
}
