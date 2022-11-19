package html

import (
	"net/http"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
	otfhttp "github.com/leg100/otf/http"
	"github.com/r3labs/sse/v2"
)

const DefaultPathPrefix = "/"

// Application is the otf web app.
type Application struct {
	// Static asset server
	staticServer http.FileSystem
	// otf service accessors
	otf.Application
	// path prefix for all URLs
	pathPrefix string
	// view engine populates and renders templates
	*viewEngine
	// logger for logging messages
	logr.Logger
	// server-side-events server
	*sse.Server
	// enabled authenticators
	authenticators []*Authenticator
	// site admin's authentication token
	siteToken string
	// secret for webhook signatures
	secret string
	// mapping of cloud name to cloud
	cloudDB cloudDB
}

type ApplicationOption func(*Application)

func WithSiteToken(token string) ApplicationOption {
	return func(app *Application) {
		app.siteToken = token
	}
}

// AddRoutes adds routes for the html web app.
//
// TODO: merge config and srvConfig
func AddRoutes(logger logr.Logger, config *Config, srvConfig *otfhttp.ServerConfig, services otf.Application, router *otfhttp.Router) error {
	if config.DevMode {
		logger.Info("enabled developer mode")
	}
	views, err := newViewEngine(config.DevMode)
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
		Application:  services,
		staticServer: newStaticServer(config.DevMode),
		pathPrefix:   DefaultPathPrefix,
		viewEngine:   views,
		Logger:       logger,
		Server:       sseServer,
		siteToken:    srvConfig.SiteToken,
		secret:       srvConfig.Secret,
		cloudDB:      config.CloudConfigs,
	}

	app.authenticators, err = newAuthenticators(services, config.CloudConfigs)
	if err != nil {
		return err
	}

	app.addRoutes(router)
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
		r.GET(auth.RequestPath(), auth.requestHandler)
		r.GET(auth.callbackPath(), auth.responseHandler)
	}
	r.GET("/admin/login", app.adminLoginPromptHandler)
	r.PST("/admin/login", app.adminLoginHandler)

	// routes that require authentication.
	r.Sub(func(r *otfhttp.Router) {
		r.Use((&authMiddleware{app, app}).authenticate)
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
		r.PST("/organizations/{organization_name}/agent-tokens/delete", app.deleteAgentToken)
		r.PST("/organizations/{organization_name}/agent-tokens/create", app.createAgentToken)
		r.GET("/organizations/{organization_name}/agent-tokens/new", app.newAgentToken)

		r.GET("/organizations/{organization_name}/vcs-providers", app.listVCSProviders)
		r.GET("/organizations/{organization_name}/vcs-providers/{cloud_name}/new", app.newVCSProvider)
		r.PST("/organizations/{organization_name}/vcs-providers/{cloud_name}/create", app.createVCSProvider)
		r.PST("/organizations/{organization_name}/vcs-providers/delete", app.deleteVCSProvider)

		r.GET("/organizations", app.listOrganizations)
		r.GET("/organizations/new", app.newOrganization)
		r.PST("/organizations/create", app.createOrganization)
		r.GET("/organizations/{organization_name}", app.getOrganization)
		r.GET("/organizations/{organization_name}/edit", app.editOrganization)
		r.PST("/organizations/{organization_name}/update", app.updateOrganization)
		r.PST("/organizations/{organization_name}/delete", app.deleteOrganization)

		r.GET("/organizations/{organization_name}/users", app.listUsers)

		r.GET("/organizations/{organization_name}/teams", app.listTeams)
		r.GET("/organizations/{organization_name}/teams/{team_name}", app.getTeam)
		r.GET("/organizations/{organization_name}/teams/{team_name}/users", app.listTeamUsers)
		r.PST("/organizations/{organization_name}/teams/{team_name}/update", app.updateTeam)

		r.GET("/organizations/{organization_name}/permissions", app.listOrganizationPermissions)

		r.GET("/organizations/{organization_name}/workspaces", app.listWorkspaces)
		r.GET("/organizations/{organization_name}/workspaces/new", app.newWorkspace)
		r.PST("/organizations/{organization_name}/workspaces/create", app.createWorkspace)
		r.GET("/organizations/{organization_name}/workspaces/{workspace_name}", app.getWorkspace)
		r.GET("/organizations/{organization_name}/workspaces/{workspace_name}/edit", app.editWorkspace)
		r.PST("/organizations/{organization_name}/workspaces/{workspace_name}/update", app.updateWorkspace)
		r.PST("/organizations/{organization_name}/workspaces/{workspace_name}/delete", app.deleteWorkspace)
		r.PST("/organizations/{organization_name}/workspaces/{workspace_name}/lock", app.lockWorkspace)
		r.PST("/organizations/{organization_name}/workspaces/{workspace_name}/unlock", app.unlockWorkspace)
		r.PST("/organizations/{organization_name}/workspaces/{workspace_name}/permissions", app.setWorkspacePermission)
		r.PST("/organizations/{organization_name}/workspaces/{workspace_name}/permissions/unset", app.unsetWorkspacePermission)
		r.GET("/organizations/{organization_name}/workspaces/{workspace_name}/vcs-providers", app.listWorkspaceVCSProviders)
		r.GET("/organizations/{organization_name}/workspaces/{workspace_name}/vcs-providers/{vcs_provider_id}/repos", app.listWorkspaceVCSRepos)
		r.PST("/organizations/{organization_name}/workspaces/{workspace_name}/vcs-providers/{vcs_provider_id}/repos/connect", app.connectWorkspaceRepo)
		r.PST("/organizations/{organization_name}/workspaces/{workspace_name}/repo/disconnect", app.disconnectWorkspaceRepo)
		r.PST("/organizations/{organization_name}/workspaces/{workspace_name}/start-run", app.startRun)

		r.GET("/organizations/{organization_name}/workspaces/{workspace_name}/watch", app.watchWorkspace)
		r.GET("/organizations/{organization_name}/workspaces/{workspace_name}/runs", app.listRuns)
		r.GET("/organizations/{organization_name}/workspaces/{workspace_name}/runs/new", app.newRun)
		r.GET("/organizations/{organization_name}/workspaces/{workspace_name}/runs/{run_id}", app.getRun)
		r.GET("/organizations/{organization_name}/workspaces/{workspace_name}/runs/{run_id}/tail", app.tailRun)
		r.PST("/organizations/{organization_name}/workspaces/{workspace_name}/runs/{run_id}/delete", app.deleteRun)
		r.PST("/organizations/{organization_name}/workspaces/{workspace_name}/runs/{run_id}/cancel", app.cancelRun)
		r.PST("/organizations/{organization_name}/workspaces/{workspace_name}/runs/{run_id}/apply", app.applyRun)
		r.PST("/organizations/{organization_name}/workspaces/{workspace_name}/runs/{run_id}/discard", app.discardRun)

		// this handles the link the terraform CLI shows during a plan/apply.
		r.GET("/app/{organization_name}/{workspace_name}/runs/{run_id}", app.getRun)

		// terraform login opens a browser to this page
		r.GET("/app/settings/tokens", app.tokensHandler)
	})
}
