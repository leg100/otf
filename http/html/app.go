package html

import (
	"net/http"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
)

const DefaultPathPrefix = "/"

// Application is the oTF web app.
type Application struct {
	// Static asset server
	staticServer http.FileSystem
	// oTF service accessors
	otf.Application
	// github oauth authorization
	oauth *githubOAuthApp
	// path prefix for all URLs
	pathPrefix string
	// view engine populates and renders templates
	*viewEngine
}

// AddRoutes adds routes for the html web app.
func AddRoutes(logger logr.Logger, config Config, services otf.Application, router *Router) error {
	if config.DevMode {
		logger.Info("enabled developer mode")
	}
	views, err := newViewEngine(config.DevMode)
	if err != nil {
		return err
	}
	oauthApp, err := newGithubOAuthApp(config.Github)
	if err != nil {
		return err
	}
	app := &Application{
		Application:  services,
		oauth:        oauthApp,
		staticServer: newStaticServer(config.DevMode),
		pathPrefix:   DefaultPathPrefix,
		viewEngine:   views,
	}
	app.addRoutes(router)
	return nil
}

// AddRoutes adds application routes and middleware to an HTTP multiplexer.
func (app *Application) addRoutes(r *Router) {
	r.Handle("/", http.RedirectHandler("/organizations", http.StatusFound))

	// Static assets (JS, CSS, etc).
	r.PathPrefix("/static/").Handler(http.FileServer(app.staticServer)).Methods("GET")

	// Redirect paths with a trailing slash to path without, e.g. /runs/ ->
	// /runs. Uses an HTTP301.
	r.StrictSlash(true)

	// routes that don't require authentication.
	r.Sub(func(r *Router) {
		r.GET("/login", app.loginHandler)
		// github routes
		r.Sub(func(r *Router) {
			r.GET("/github/login", app.oauth.requestHandler)
			r.GET(githubCallbackPath, app.githubLogin)
		})
	})
	// routes that require authentication.
	r.Sub(func(r *Router) {
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

		r.GET("/organizations", app.listOrganizations)
		r.GET("/organizations/{organization_name}", app.getOrganization)
		r.GET("/organizations/{organization_name}/overview", app.getOrganizationOverview)
		r.GET("/organizations/{organization_name}/edit", app.editOrganization)
		r.PST("/organizations/{organization_name}/update", app.updateOrganization)
		r.PST("/organizations/{organization_name}/delete", app.deleteOrganization)

		r.GET("/organizations/{organization_name}/workspaces", app.listWorkspaces)
		r.GET("/organizations/{organization_name}/workspaces/new", app.newWorkspace)
		r.PST("/organizations/{organization_name}/workspaces/create", app.createWorkspace)
		r.GET("/organizations/{organization_name}/workspaces/{workspace_name}", app.getWorkspace)
		r.GET("/organizations/{organization_name}/workspaces/{workspace_name}/edit", app.editWorkspace)
		r.PST("/organizations/{organization_name}/workspaces/{workspace_name}/update", app.updateWorkspace)
		r.PST("/organizations/{organization_name}/workspaces/{workspace_name}/delete", app.deleteWorkspace)
		r.PST("/organizations/{organization_name}/workspaces/{workspace_name}/lock", app.lockWorkspace)
		r.PST("/organizations/{organization_name}/workspaces/{workspace_name}/unlock", app.unlockWorkspace)

		r.GET("/organizations/{organization_name}/workspaces/{workspace_name}/runs", app.listRuns)
		r.GET("/organizations/{organization_name}/workspaces/{workspace_name}/runs/new", app.newRun)
		r.PST("/organizations/{organization_name}/workspaces/{workspace_name}/runs/create", app.createRun)
		r.GET("/organizations/{organization_name}/workspaces/{workspace_name}/runs/{run_id}", app.getRun)
		r.GET("/organizations/{organization_name}/workspaces/{workspace_name}/runs/{run_id}/plan", app.getPlan)
		r.GET("/organizations/{organization_name}/workspaces/{workspace_name}/runs/{run_id}/apply", app.getApply)
		r.PST("/organizations/{organization_name}/workspaces/{workspace_name}/runs/{run_id}/delete", app.deleteRun)
		r.PST("/organizations/{organization_name}/workspaces/{workspace_name}/runs/{run_id}/cancel", app.cancelRun)

		// this handles the link the terraform CLI shows during a plan/apply.
		r.GET("/app/{organization_name}/{workspace_name}/runs/{run_id}", app.getRun)

		// terraform login opens a browser to this page
		r.GET("/app/settings/tokens", app.tokensHandler)
	})
}
