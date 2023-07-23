// Package api provides http handlers for the API.
package api

import (
	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/auth"
	"github.com/leg100/otf/internal/configversion"
	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/notifications"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/run"
	"github.com/leg100/otf/internal/state"
	"github.com/leg100/otf/internal/tokens"
	"github.com/leg100/otf/internal/variable"
	"github.com/leg100/otf/internal/vcsprovider"
	"github.com/leg100/otf/internal/workspace"
	"github.com/leg100/surl"
)

type (
	api struct {
		logr.Logger

		run.RunService
		organization.OrganizationService
		state.StateService
		workspace.WorkspaceService
		configversion.ConfigurationVersionService
		auth.AuthService
		tokens.TokensService
		variable.VariableService
		notifications.NotificationService
		vcsprovider.VCSProviderService

		marshaler
		internal.Verifier // for verifying signed urls

		maxConfigSize int64 // Maximum permitted config upload size in bytes
	}

	Options struct {
		run.RunService
		organization.OrganizationService
		state.StateService
		workspace.WorkspaceService
		configversion.ConfigurationVersionService
		auth.AuthService
		auth.TeamService
		tokens.TokensService
		variable.VariableService
		notifications.NotificationService
		vcsprovider.VCSProviderService

		*surl.Signer

		MaxConfigSize int64
	}
)

func New(opts Options) *api {
	return &api{
		OrganizationService:         opts.OrganizationService,
		WorkspaceService:            opts.WorkspaceService,
		RunService:                  opts.RunService,
		StateService:                opts.StateService,
		ConfigurationVersionService: opts.ConfigurationVersionService,
		AuthService:                 opts.AuthService,
		Verifier:                    opts.Signer,
		TokensService:               opts.TokensService,
		VariableService:             opts.VariableService,
		NotificationService:         opts.NotificationService,
		VCSProviderService:          opts.VCSProviderService,
		marshaler: &jsonapiMarshaler{
			OrganizationService: opts.OrganizationService,
			WorkspaceService:    opts.WorkspaceService,
			RunService:          opts.RunService,
			StateService:        opts.StateService,
			TeamService:         opts.TeamService,
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
	a.addTagHandlers(r)
	a.addConfigHandlers(r)
	a.addUserHandlers(r)
	a.addTeamHandlers(r)
	a.addTeamMembershipHandlers(r)
	a.addVariableHandlers(r)
	a.addTokenHandlers(r)
	a.addNotificationHandlers(r)
	a.addOrganizationMembershipHandlers(r)
	a.addOAuthClientHandlers(r)
}
