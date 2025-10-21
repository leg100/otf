package ui

import (
	"encoding/json"

	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/authenticator"
	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/engine"
	"github.com/leg100/otf/internal/github"
	"github.com/leg100/otf/internal/http/html/paths"
	"github.com/leg100/otf/internal/module"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/run"
	"github.com/leg100/otf/internal/runner"
	"github.com/leg100/otf/internal/state"
	"github.com/leg100/otf/internal/team"
	"github.com/leg100/otf/internal/tokens"
	"github.com/leg100/otf/internal/user"
	"github.com/leg100/otf/internal/variable"
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
	State                        *state.Service
	Runners                      *runner.Service
	GithubApp                    *github.Service
	EngineService                *engine.Service
	HostnameService              *internal.HostnameService
	Tokens                       *tokens.Service
	Authorizer                   *authz.Authorizer
	AuthenticatorService         *authenticator.Service
	VariablesService             *variable.Service
	GithubHostname               *internal.WebURL
	SkipTLSVerification          bool
	SiteToken                    string
	RestrictOrganizationCreation bool
}

// AddHandlers registers all UI handlers with the router
func (h *Handlers) AddHandlers(r *mux.Router) {
	// Unauthenticated, unprefixed routes
	addLoginHandlers(r, h.AuthenticatorService, h.Tokens, h.SiteToken)

	// Add UI prefix to paths handled by handlers below.
	// NOTE: all UI prefixed paths are intercepted by middleware that mandates
	// the request is authenticated.
	r = r.PathPrefix(paths.UIPrefix).Subrouter()

	addRunHandlers(r, h.Logger, h.Runs, h.Workspaces, h.Users, h.Runs)
	addTeamHandlers(r, h.Teams, h.Users, h.Tokens, h.Authorizer)
	addUserHandlers(r, h.Users, h.Authorizer)
	addWorkspaceHandlers(r, h.Logger, h.Workspaces, h.Teams, h.VCSProviders, h.Authorizer, h.EngineService, h.Runs, h.Users)
	addOrganizationHandlers(r, h.Organizations, h.RestrictOrganizationCreation)
	addModuleHandlers(r, h.Modules, h.VCSProviders, h.HostnameService, h.Authorizer)
	addVariableHandlers(r, h.VariablesService, h.Workspaces, h.Authorizer)
	addRunnerHandlers(r, h.Runners, h.Workspaces, h.Authorizer, h.Logger)
	addStateHandlers(r, h.State)
	addVCSHandlers(r, h.VCSProviders, h.HostnameService)

	githubHandlers := &githubHandlers{
		HostnameService:     h.HostnameService,
		svc:                 h.GithubApp,
		authorizer:          h.Authorizer,
		githubAPIURL:        h.GithubHostname,
		skipTLSVerification: h.SkipTLSVerification,
	}
	githubHandlers.addHandlers(r)
}

func toJSON(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}
