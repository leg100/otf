// Package api provides http handlers for the API.
package api

import (
	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/leg100/otf/logr"
	"github.com/leg100/otf/organization"
	"github.com/leg100/otf/orgcreator"
	"github.com/leg100/otf/run"
	"github.com/leg100/otf/state"
	"github.com/leg100/otf/workspace"
)

type (
	api struct {
		logr.Logger

		run.RunService
		organization.OrganizationService
		orgcreator.OrganizationCreatorService
		state.StateService
		workspace.WorkspaceService

		marshaler
	}

	Options struct {
		run.RunService
		organization.OrganizationService
		orgcreator.OrganizationCreatorService
		state.StateService
		workspace.WorkspaceService
		otf.Signer
	}
)

func New(opts Options) *api {
	return &api{
		OrganizationService:        opts.OrganizationService,
		OrganizationCreatorService: opts.OrganizationCreatorService,
		WorkspaceService:           opts.WorkspaceService,
		RunService:                 opts.RunService,
		StateService:               opts.StateService,
		marshaler: &jsonapiMarshaler{
			OrganizationService: opts.OrganizationService,
			WorkspaceService:    opts.WorkspaceService,
			RunService:          opts.RunService,
			StateService:        opts.StateService,
			runLogsURLGenerator: &runLogsURLGenerator{opts.Signer},
		},
	}
}

func (a *api) AddHandlers(r *mux.Router) {
	a.addOrganizationHandlers(r)
	a.addRunHandlers(r)
	a.addWorkspaceHandlers(r)
	a.addStateHandlers(r)
}
