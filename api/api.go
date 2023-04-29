// Package api provides http handlers for the API.
package api

import (
	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/leg100/otf/auth"
	"github.com/leg100/otf/configversion"
	"github.com/leg100/otf/logr"
	"github.com/leg100/otf/organization"
	"github.com/leg100/otf/orgcreator"
	"github.com/leg100/otf/run"
	"github.com/leg100/otf/state"
	"github.com/leg100/otf/tokens"
	"github.com/leg100/otf/variable"
	"github.com/leg100/otf/workspace"
	"github.com/leg100/surl"
)

type (
	api struct {
		logr.Logger

		run.RunService
		organization.OrganizationService
		orgcreator.OrganizationCreatorService
		state.StateService
		workspace.WorkspaceService
		configversion.ConfigurationVersionService
		auth.AuthService
		tokens.TokensService
		variable.VariableService

		marshaler
		otf.Verifier // for verifying signed urls

		maxConfigSize int64 // Maximum permitted config upload size in bytes
	}

	Options struct {
		run.RunService
		organization.OrganizationService
		orgcreator.OrganizationCreatorService
		state.StateService
		workspace.WorkspaceService
		configversion.ConfigurationVersionService
		auth.AuthService
		tokens.TokensService
		variable.VariableService

		*surl.Signer

		MaxConfigSize int64
	}
)

func New(opts Options) *api {
	return &api{
		OrganizationService:         opts.OrganizationService,
		OrganizationCreatorService:  opts.OrganizationCreatorService,
		WorkspaceService:            opts.WorkspaceService,
		RunService:                  opts.RunService,
		StateService:                opts.StateService,
		ConfigurationVersionService: opts.ConfigurationVersionService,
		AuthService:                 opts.AuthService,
		Verifier:                    opts.Signer,
		TokensService:               opts.TokensService,
		VariableService:             opts.VariableService,
		marshaler: &jsonapiMarshaler{
			OrganizationService: opts.OrganizationService,
			WorkspaceService:    opts.WorkspaceService,
			RunService:          opts.RunService,
			StateService:        opts.StateService,
			runLogsURLGenerator: &runLogsURLGenerator{opts.Signer},
		},
		maxConfigSize: opts.MaxConfigSize,
	}
}

func (a *api) AddHandlers(r *mux.Router) {
	a.addOrganizationHandlers(r)
	a.addRunHandlers(r)
	a.addWorkspaceHandlers(r)
	a.addStateHandlers(r)
	a.addConfigHandlers(r)
	a.addUserHandlers(r)
	a.addTeamHandlers(r)
	a.addVariableHandlers(r)
	a.addTokenHandlers(r)
}
