package html

import (
	"net/http"

	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"github.com/leg100/otf"
)

const DefaultPathPrefix = "/"

// Application is the oTF web app.
type Application struct {
	// HTML template renderer
	renderer
	// Static asset server
	staticServer http.FileSystem
	// oTF service accessors
	otf.Application
	// github oauth authorization
	oauth *githubOAuthApp
	// path prefix for all URLs
	pathPrefix string
	// factory for making templateData structs
	*viewEngine
	// wrapper around mux router
	*router
}

// AddRoutes adds routes for the html web app.
func AddRoutes(logger logr.Logger, config Config, services otf.Application, muxrouter *mux.Router) error {
	if config.DevMode {
		logger.Info("enabled developer mode")
	}
	views, err := newViewEngine(&router{muxrouter}, config.DevMode)
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
		router:       &router{Router: muxrouter},
	}
	app.addRoutes(app.router)
	return nil
}

// AddRoutes adds application routes and middleware to an HTTP multiplexer.
func (app *Application) addRoutes(r *router) {
	r.Handle("/", http.RedirectHandler("/organizations", http.StatusFound))

	// Static assets (JS, CSS, etc).
	r.PathPrefix("/static/").Handler(http.FileServer(app.staticServer)).Methods("GET")

	// Redirect paths with a trailing slash to path without, e.g. /runs/ ->
	// /runs. Uses an HTTP301.
	r.StrictSlash(true)

	// routes that don't require authentication.
	r.sub(func(r *router) {
		r.get("/login", app.loginHandler).Name("login")
		// github routes
		r.sub(func(r *router) {
			r.get("/github/login", app.oauth.requestHandler)
			r.get(githubCallbackPath, app.githubLogin)
		})
	})
	// routes that require authentication.
	r.sub(func(r *router) {
		r.Use(app.authenticateUser)
		r.Use(app.setCurrentOrganization)

		r.pst("/logout", app.logoutHandler).Name("logout")
		r.get("/profile", app.profileHandler).Name("getProfile")
		r.get("/profile/sessions", app.sessionsHandler).Name("listSession")
		r.pst("/profile/sessions/revoke", app.revokeSessionHandler).Name("revokeSession")

		r.get("/profile/tokens", app.tokensHandler).Name("listToken")
		r.pst("/profile/tokens/delete", app.deleteTokenHandler).Name("deleteToken")
		r.get("/profile/tokens/new", app.newTokenHandler).Name("newToken")
		r.pst("/profile/tokens/create", app.createTokenHandler).Name("createToken")

		r.get("/organizations/", app.listOrganizations).Name("listOrganization")
		r.get("/organizations/new", app.newOrganization).Name("newOrganization")
		r.pst("/organizations/create", app.createOrganization).Name("createOrganization")
		r.get("/organizations/{organization_name}", app.getOrganization).Name("getOrganization")
		r.get("/organizations/{organization_name}/overview", app.getOrganizationOverview).Name("getOrganizationOverview")
		r.get("/organizations/{organization_name}/edit", app.editOrganization).Name("editOrganization")
		r.pst("/organizations/{organization_name}/update", app.updateOrganization).Name("updateOrganization")
		r.pst("/organizations/{organization_name}/delete", app.deleteOrganization).Name("deleteOrganization")

		r.get("/organizations/{organization_name}/workspaces", app.listWorkspaces).Name("listWorkspace")
		r.get("/organizations/{organization_name}/workspaces/new", app.newWorkspace).Name("newWorkspace")
		r.pst("/organizations/{organization_name}/workspaces/create", app.createWorkspace).Name("createWorkspace")
		r.get("/organizations/{organization_name}/workspaces/{workspace_name}", app.getWorkspace).Name("getWorkspace")
		r.get("/organizations/{organization_name}/workspaces/{workspace_name}/edit", app.editWorkspace).Name("editWorkspace")
		r.pst("/organizations/{organization_name}/workspaces/{workspace_name}/update", app.updateWorkspace).Name("updateWorkspace")
		r.pst("/organizations/{organization_name}/workspaces/{workspace_name}/delete", app.deleteWorkspace).Name("deleteWorkspace")
		r.pst("/organizations/{organization_name}/workspaces/{workspace_name}/lock", app.lockWorkspace).Name("lockWorkspace")
		r.pst("/organizations/{organization_name}/workspaces/{workspace_name}/unlock", app.unlockWorkspace).Name("unlockWorkspace")

		r.get("/organizations/{organization_name}/workspaces/{workspace_name}/runs", app.listRuns).Name("listRun")
		r.get("/organizations/{organization_name}/workspaces/{workspace_name}/runs/new", app.newRun).Name("newRun")
		r.pst("/organizations/{organization_name}/workspaces/{workspace_name}/runs/create", app.createRun).Name("createRun")
		r.get("/organizations/{organization_name}/workspaces/{workspace_name}/runs/{run_id}", app.getRun).Name("getRun")
		r.get("/organizations/{organization_name}/workspaces/{workspace_name}/runs/{run_id}/plan", app.getPlan).Name("getPlan")
		r.get("/organizations/{organization_name}/workspaces/{workspace_name}/runs/{run_id}/apply", app.getApply).Name("getApply")
		r.pst("/organizations/{organization_name}/workspaces/{workspace_name}/runs/{run_id}/delete", app.deleteRun).Name("deleteRun")
	})
}
