package watch

import (
	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/r3labs/sse/v2"
)

type service struct {
	otf.Authorizer
	otf.PubSubService
	logr.Logger

	*app

	api *api
	web *web
}

func NewService(opts Options) *service {
	app := &app{
		Authorizer:    opts.Authorizer,
		Logger:        opts.Logger,
		PubSubService: opts.PubSubService,
	}

	// Create and configure SSE server
	srv := newSSEServer()

	return &service{
		api: &api{
			app:          app,
			eventsServer: srv,
		},
		app: app,
		web: &web{
			Renderer: opts.Renderer,
			app:      app,
			Server:   srv,
		},
	}
}

type Options struct {
	otf.Authorizer
	otf.PubSubService
	otf.Renderer
	logr.Logger
}

func (s *service) AddHandlers(r *mux.Router) {
	s.api.addHandlers(r)
	s.web.addHandlers(r)
}

func newSSEServer() *sse.Server {
	// Create and configure SSE server
	srv := sse.New()
	// we don't use last-event-item functionality so turn it off
	srv.AutoReplay = false
	// encode payloads into base64 otherwise the JSON string payloads corrupt
	// the SSE protocol
	srv.EncodeBase64 = true

	return srv
}
