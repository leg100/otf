package ui

import (
	"encoding/json"

	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/authenticator"
	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/engine"
	"github.com/leg100/otf/internal/module"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/run"
	"github.com/leg100/otf/internal/team"
	"github.com/leg100/otf/internal/tokens"
	"github.com/leg100/otf/internal/user"
	"github.com/leg100/otf/internal/vcs"
	"github.com/leg100/otf/internal/workspace"
)

// Handlers registers all UI handlers
type Handlers struct {
	Logger                       logr.Logger
	Runs                         *run.Service
	Workspaces                   *workspace.Service
	Users                        *user.Service
	Teams                        *team.Service
	Organizations                *organization.Service
	Modules                      *module.Service
	VCSProviders                 *vcs.Service
	EngineService                *engine.Service
	HostnameService              *internal.HostnameService
	Tokens                       *tokens.Service
	Authorizer                   *authz.Authorizer
	AuthenticatorService         *authenticator.Service
	SiteToken                    string
	RestrictOrganizationCreation bool
}

// AddHandlers registers all UI handlers with the router
func (h *Handlers) AddHandlers(r *mux.Router) {
	AddRunHandlers(r, h.Logger, h.Runs, h.Workspaces, h.Users, h.Runs)
	AddTeamHandlers(r, h.Teams, h.Tokens, h.Teams)
	AddUserHandlers(r, h.Users, h.Teams, h.Tokens, h.Authorizer, h.SiteToken)
	AddWorkspaceHandlers(r, h.Logger, h.Workspaces, h.Teams, h.VCSProviders, h.Authorizer, h.EngineService)
	AddOrganizationHandlers(r, h.Organizations, h.RestrictOrganizationCreation)
	AddModuleHandlers(r, h.Modules, h.VCSProviders, h.HostnameService, h.Authorizer)
	AddModuleHandlers(r, h.Modules, h.VCSProviders, h.HostnameService, h.Authorizer)
	addLoginHandlers(r, h.AuthenticatorService)
}

func toJSON(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}
