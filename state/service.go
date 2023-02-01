package state

import (
	"github.com/go-logr/logr"
	"github.com/leg100/otf"
)

var _ otf.StateVersionService = (*service)(nil)

// service provides access to state and state versions
type service struct {
	*app
	*handlers
}

func NewService(opts ServiceOptions) *service {
	app := &app{
		Authorizer: opts.Authorizer,
		Logger:     opts.Logger,
		db:         newPGDB(opts.Database),
		cache:      opts.Cache,
	}
	return &service{
		app:      app,
		handlers: &handlers{app},
	}
}

type ServiceOptions struct {
	otf.Authorizer
	otf.Database
	otf.Cache
	logr.Logger
}
