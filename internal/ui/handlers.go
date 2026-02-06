package ui

import (
	"encoding/json"
	"net/http"

	"github.com/a-h/templ"
	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/authenticator"
	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/configversion"
	"github.com/leg100/otf/internal/engine"
	"github.com/leg100/otf/internal/github"
	"github.com/leg100/otf/internal/http/html"
	"github.com/leg100/otf/internal/http/html/paths"
	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/module"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/run"
	"github.com/leg100/otf/internal/runner"
	"github.com/leg100/otf/internal/state"
	"github.com/leg100/otf/internal/team"
	"github.com/leg100/otf/internal/tokens"
	"github.com/leg100/otf/internal/ui/helpers"
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
	Configs                      *configversion.Service
	HostnameService              *internal.HostnameService
	Tokens                       *tokens.Service
	Authorizer                   *authz.Authorizer
	AuthenticatorService         *authenticator.Service
	VariablesService             *variable.Service
	GithubHostname               *internal.WebURL
	SkipTLSVerification          bool
	SiteToken                    string
	RestrictOrganizationCreation bool

	templates *templates
}

func NewHandlers(
	Logger logr.Logger,
	Runs *run.Service,
	Workspaces *workspace.Service,
	Users *user.Service,
	Teams *team.Service,
	Organizations *organization.Service,
	Modules *module.Service,
	VCSProviders *vcs.Service,
	State *state.Service,
	Runners *runner.Service,
	GithubApp *github.Service,
	EngineService *engine.Service,
	Configs *configversion.Service,
	HostnameService *internal.HostnameService,
	Tokens *tokens.Service,
	Authorizer *authz.Authorizer,
	AuthenticatorService *authenticator.Service,
	VariablesService *variable.Service,
	GithubHostname *internal.WebURL,
	SkipTLSVerification bool,
	SiteToken string,
	RestrictOrganizationCreation bool,
) *Handlers {
	return &Handlers{
		Logger:                       Logger,
		Runs:                         Runs,
		Workspaces:                   Workspaces,
		Users:                        Users,
		Teams:                        Teams,
		Organizations:                Organizations,
		Modules:                      Modules,
		VCSProviders:                 VCSProviders,
		State:                        State,
		Runners:                      Runners,
		GithubApp:                    GithubApp,
		EngineService:                EngineService,
		Configs:                      Configs,
		HostnameService:              HostnameService,
		Tokens:                       Tokens,
		Authorizer:                   Authorizer,
		AuthenticatorService:         AuthenticatorService,
		VariablesService:             VariablesService,
		GithubHostname:               GithubHostname,
		SkipTLSVerification:          SkipTLSVerification,
		SiteToken:                    SiteToken,
		RestrictOrganizationCreation: RestrictOrganizationCreation,
		templates: &templates{
			configs:    Configs,
			workspaces: Workspaces,
			users:      Users,
		},
	}
}

// AddHandlers registers all UI handlers with the router
func (h *Handlers) AddHandlers(r *mux.Router) {
	// Unauthenticated, unprefixed routes
	addLoginHandlers(r, h)

	// Add UI prefix to paths handled by handlers below.
	// NOTE: all UI prefixed paths are intercepted by middleware that mandates
	// the request is authenticated.
	r = r.PathPrefix(paths.UIPrefix).Subrouter()

	addRunHandlers(r, h)
	addTeamHandlers(r, h)
	addUserHandlers(r, h)
	addWorkspaceHandlers(r, h)
	addOrganizationHandlers(r, h)
	addModuleHandlers(r, h)
	addVariableHandlers(r, h)
	addRunnerHandlers(r, h)
	addStateHandlers(r, h)
	addVCSHandlers(r, h)
	addGithubAppHandlers(r, h)
}

func (h *Handlers) renderPage(comp templ.Component, title string, w http.ResponseWriter, r *http.Request, opts ...renderPageOption) {
	var props helpers.LayoutProps
	for _, o := range opts {
		o(&props)
	}
	// Render component as a child of the layout component.
	layout := helpers.Layout(props)
	html.Render(layout, w, r, html.WithChildren(comp))
}

type renderPageOption func(opts *helpers.LayoutProps)

func withOrganization(org resource.ID) renderPageOption {
	return func(opts *helpers.LayoutProps) {
		opts.Organization = org
	}
}

func withWorkspace(ws *workspace.Workspace) renderPageOption {
	return func(opts *helpers.LayoutProps) {
		opts.Organization = internal.Ptr(ws.Organization)
		opts.Workspace = internal.Ptr(ws.Info())
	}
}

func withBreadcrumbs(crumbs ...helpers.Breadcrumb) renderPageOption {
	return func(opts *helpers.LayoutProps) {
		opts.Breadcrumbs = append(opts.Breadcrumbs, crumbs...)
	}
}

func withContentLinks(comp templ.Component) renderPageOption {
	return func(opts *helpers.LayoutProps) {
		opts.ContentLinks = comp
	}
}

func withContentActions(comp templ.Component) renderPageOption {
	return func(opts *helpers.LayoutProps) {
		opts.ContentActions = comp
	}
}

func withPreContent(comp templ.Component) renderPageOption {
	return func(opts *helpers.LayoutProps) {
		opts.PreContent = comp
	}
}

func withPostContent(comp templ.Component) renderPageOption {
	return func(opts *helpers.LayoutProps) {
		opts.PostContent = comp
	}
}

func toJSON(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}
