package html

import (
	"net/http"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
	"github.com/leg100/otf/cloud"
	otfhttp "github.com/leg100/otf/http"
	"github.com/r3labs/sse/v2"
)

const DefaultPathPrefix = "/"

// Application is the otf web app.
type Application struct {
	staticServer    http.FileSystem  // Static asset server
	pathPrefix      string           // path prefix for all URLs
	authenticators  []*Authenticator // enabled authenticators
	siteToken       string           // site admin's authentication token
	secret          string           // secret for webhook signatures
	variableService otf.VariableService

	otf.Application // otf service accessors
	*viewEngine     // view engine populates and renders templates
	logr.Logger     // logger for logging messages
	*sse.Server     // server-side-events server
}

// ApplicationOptions are options for configuring the web app
type ApplicationOptions struct {
	DevMode      bool
	CloudConfigs []*cloud.CloudOAuthConfig // for configuring authenticators

	*otfhttp.ServerConfig
	*otfhttp.Router
	otf.Application
	otf.VariableService
}

// AddRoutes adds routes for the html web app.
func AddRoutes(logger logr.Logger, opts ApplicationOptions) error {
	logger = logger.WithValues("component", "html")
	if opts.DevMode {
		logger.Info("enabled developer mode")
	}
	views, err := newViewEngine(viewEngineOptions{
		devMode: opts.DevMode,
	})
	if err != nil {
		return err
	}

	// Setup SSE server
	sseServer := sse.New()
	// delegate responsibility to sse lib to create/delete streams
	sseServer.AutoStream = true
	// we don't use last-event-item functionality so turn it off
	sseServer.AutoReplay = false

	app := &Application{
		Application:     opts.Application,
		variableService: opts.VariableService,
		staticServer:    newStaticServer(opts.DevMode),
		pathPrefix:      DefaultPathPrefix,
		viewEngine:      views,
		Logger:          logger,
		Server:          sseServer,
		siteToken:       opts.SiteToken,
		secret:          opts.Secret,
	}

	app.authenticators, err = newAuthenticators(logger, opts.Application, opts.CloudConfigs)
	if err != nil {
		return err
	}

	app.addRoutes(opts.Router)
	return nil
}

// AddRoutes adds application routes and middleware to an HTTP multiplexer.
func (app *Application) addRoutes(r *otfhttp.Router) {
	r.Handle("/", http.RedirectHandler("/organizations", http.StatusFound))

	// Static assets (JS, CSS, etc).
	r.PathPrefix("/static/").Handler(http.FileServer(app.staticServer)).Methods("GET")

	// Redirect paths with a trailing slash to path without, e.g. /runs/ ->
	// /runs. Uses an HTTP301.
	r.StrictSlash(true)

	// routes that don't require authentication.
	r.GET("/login", app.loginHandler)
	for _, auth := range app.authenticators {
		r.GET(auth.RequestPath(), auth.RequestHandler)
		r.GET(auth.CallbackPath(), auth.responseHandler)
	}
	r.GET("/admin/login", app.adminLoginPromptHandler)
	r.PST("/admin/login", app.adminLoginHandler)

	// routes that require authentication.
	r.Sub(func(r *otfhttp.Router) {
		r.Use((&authMiddleware{app}).authenticate)
		r.Use(setOrganization)

		r.PST("/logout", app.logoutHandler)
		r.GET("/profile", app.profileHandler)
		r.GET("/profile/sessions", app.sessionsHandler)
		r.PST("/profile/sessions/revoke", app.revokeSessionHandler)

		r.GET("/profile/tokens", app.tokensHandler)
		r.PST("/profile/tokens/delete", app.deleteTokenHandler)
		r.GET("/profile/tokens/new", app.newTokenHandler)
		r.PST("/profile/tokens/create", app.createTokenHandler)

		r.GET("/organizations/{organization_name}/agent-tokens", app.listAgentTokens)
		r.PST("/organizations/{organization_name}/agent-tokens/create", app.createAgentToken)
		r.GET("/organizations/{organization_name}/agent-tokens/new", app.newAgentToken)
		r.PST("/agent-tokens/{agent_token_id}/delete", app.deleteAgentToken)

		r.GET("/organizations/{organization_name}/vcs-providers", app.listVCSProviders)
		r.GET("/organizations/{organization_name}/vcs-providers/new", app.newVCSProvider)
		r.PST("/organizations/{organization_name}/vcs-providers/create", app.createVCSProvider)
		r.PST("/vcs-providers/{vcs_provider_id}/delete", app.deleteVCSProvider)

		r.GET("/organizations/{organization_name}/modules", app.listModules)
		r.GET("/organizations/{organization_name}/modules/new", app.newModule)
		r.GET("/organizations/{organization_name}/modules/create", app.createModule)
		r.GET("/modules/{module_id}", app.getModule)
		r.PST("/modules/{module_id}/delete", app.deleteModule)

		r.GET("/organizations", app.listOrganizations)
		r.GET("/organizations/new", app.newOrganization)
		r.PST("/organizations/create", app.createOrganization)
		r.GET("/organizations/{organization_name}", app.getOrganization)
		r.GET("/organizations/{organization_name}/edit", app.editOrganization)
		r.PST("/organizations/{organization_name}/update", app.updateOrganization)
		r.PST("/organizations/{organization_name}/delete", app.deleteOrganization)

		r.GET("/organizations/{organization_name}/users", app.listUsers)

		r.GET("/organizations/{organization_name}/teams", app.listTeams)
		r.GET("/teams/{team_id}", app.getTeam)
		r.PST("/teams/{team_id}/update", app.updateTeam)

		r.GET("/organizations/{organization_name}/permissions", app.listOrganizationPermissions)

		r.GET("/organizations/{organization_name}/workspaces", app.listWorkspaces)
		r.GET("/organizations/{organization_name}/workspaces/new", app.newWorkspace)
		r.PST("/organizations/{organization_name}/workspaces/create", app.createWorkspace)
		r.GET("/organizations/{organization_name}/workspaces/{workspace_name}", app.getWorkspaceByName)
		r.GET("/workspaces/{workspace_id}", app.getWorkspace)
		r.GET("/workspaces/{workspace_id}/edit", app.editWorkspace)
		r.PST("/workspaces/{workspace_id}/update", app.updateWorkspace)
		r.PST("/workspaces/{workspace_id}/delete", app.deleteWorkspace)
		r.PST("/workspaces/{workspace_id}/lock", app.lockWorkspace)
		r.PST("/workspaces/{workspace_id}/unlock", app.unlockWorkspace)
		r.PST("/workspaces/{workspace_id}/set-permission", app.setWorkspacePermission)
		r.PST("/workspaces/{workspace_id}/unset-permission", app.unsetWorkspacePermission)
		r.GET("/workspaces/{workspace_id}/setup-connection-provider", app.listWorkspaceVCSProviders)
		r.GET("/workspaces/{workspace_id}/setup-connection-repo", app.listWorkspaceVCSRepos)
		r.PST("/workspaces/{workspace_id}/connect", app.connectWorkspace)
		r.PST("/workspaces/{workspace_id}/disconnect", app.disconnectWorkspace)
		r.PST("/workspaces/{workspace_id}/start-run", app.startRun)

		// Variables routes
		app.variableService.AddHandlers(r.Router)

		r.GET("/workspaces/{workspace_id}/watch", app.watchWorkspace)
		r.GET("/workspaces/{workspace_id}/runs", app.listRuns)
		r.GET("/runs/{run_id}", app.getRun)
		r.GET("/runs/{run_id}/tail", app.tailRun)
		r.PST("/runs/{run_id}/delete", app.deleteRun)
		r.PST("/runs/{run_id}/cancel", app.cancelRun)
		r.PST("/runs/{run_id}/apply", app.applyRun)
		r.PST("/runs/{run_id}/discard", app.discardRun)

		// this handles the link the terraform CLI shows during a plan/apply.
		r.GET("/app/{organization_name}/{workspace_id}/runs/{run_id}", app.getRun)

		// terraform login opens a browser to this page
		r.GET("/app/settings/tokens", app.tokensHandler)
	})
}
