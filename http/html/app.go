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

	app.addRoutes(muxrouter)

	return nil
}

// AddRoutes adds application routes and middleware to an HTTP multiplexer.
func (app *Application) addRoutes(router *mux.Router) {
	router.Handle("/", http.RedirectHandler("/organizations", http.StatusFound))

	// Static assets (JS, CSS, etc).
	router.PathPrefix("/static/").Handler(http.FileServer(app.staticServer)).Methods("GET")

	// Redirect paths with a trailing slash to path without, e.g. /runs/ ->
	// /runs. Uses an HTTP301.
	router.StrictSlash(true)

	app.nonAuthRoutes(router.NewRoute().Subrouter())
	app.authRoutes(router.NewRoute().Subrouter())
}

// nonAuthRoutes adds routes that don't require authentication.
func (app *Application) nonAuthRoutes(router *mux.Router) {
	app.githubRoutes(router.NewRoute().Subrouter())

	router.HandleFunc("/login", app.loginHandler).Methods("GET").Name("login")
}

func (app *Application) githubRoutes(router *mux.Router) {
	router.HandleFunc("/github/login", app.oauth.requestHandler)
	router.HandleFunc(githubCallbackPath, app.githubLogin)
}

// authRoutes adds routes that require authentication.
func (app *Application) authRoutes(r *mux.Router) {
	r.Use(app.authenticateUser)
	r.Use(app.setCurrentOrganization)

	r.HandleFunc("/logout", app.logoutHandler).Methods("POST").Name("logout")
	r.HandleFunc("/profile", app.profileHandler).Methods("GET").Name("getProfile")
	r.HandleFunc("/profile/sessions", app.sessionsHandler).Methods("GET").Name("listSession")
	r.HandleFunc("/profile/sessions/revoke", app.revokeSessionHandler).Methods("POST").Name("revokeSession")

	r.HandleFunc("/profile/tokens", app.tokensHandler).Methods("GET").Name("listToken")
	r.HandleFunc("/profile/tokens/delete", app.deleteTokenHandler).Methods("POST").Name("deleteToken")
	r.HandleFunc("/profile/tokens/new", app.newTokenHandler).Methods("GET").Name("newToken")
	r.HandleFunc("/profile/tokens/create", app.createTokenHandler).Methods("POST").Name("createToken")

	r.HandleFunc("/organizations/", app.listOrganizations).Methods("GET").Name("listOrganization")
	r.HandleFunc("/organizations/new", app.newOrganization).Methods("GET").Name("newOrganization")
	r.HandleFunc("/organizations/create", app.createOrganization).Methods("POST").Name("createOrganization")
	r.HandleFunc("/organizations/{organization_name}", app.getOrganization).Methods("GET").Name("getOrganization")
	r.HandleFunc("/organizations/{organization_name}/overview", app.getOrganizationOverview).Methods("GET").Name("getOrganizationOverview")
	r.HandleFunc("/organizations/{organization_name}/edit", app.editOrganization).Methods("GET").Name("editOrganization")
	r.HandleFunc("/organizations/{organization_name}/update", app.updateOrganization).Methods("POST").Name("updateOrganization")
	r.HandleFunc("/organizations/{organization_name}/delete", app.deleteOrganization).Methods("POST").Name("deleteOrganization")

	r.HandleFunc("/organizations/{organization_name}/workspaces", app.listWorkspaces).Methods("GET").Name("listWorkspace")
	r.HandleFunc("/organizations/{organization_name}/workspaces/new", app.newWorkspace).Methods("GET").Name("newWorkspace")
	r.HandleFunc("/organizations/{organization_name}/workspaces/create", app.createWorkspace).Methods("POST").Name("createWorkspace")
	r.HandleFunc("/organizations/{organization_name}/workspaces/{workspace_name}", app.getWorkspace).Methods("GET").Name("getWorkspace")
	r.HandleFunc("/organizations/{organization_name}/workspaces/{workspace_name}/edit", app.editWorkspace).Methods("GET").Name("editWorkspace")
	r.HandleFunc("/organizations/{organization_name}/workspaces/{workspace_name}/update", app.updateWorkspace).Methods("POST").Name("updateWorkspace")
	r.HandleFunc("/organizations/{organization_name}/workspaces/{workspace_name}/delete", app.deleteWorkspace).Methods("POST").Name("deleteWorkspace")
	r.HandleFunc("/organizations/{organization_name}/workspaces/{workspace_name}/lock", app.lockWorkspace).Methods("POST").Name("lockWorkspace")
	r.HandleFunc("/organizations/{organization_name}/workspaces/{workspace_name}/unlock", app.unlockWorkspace).Methods("POST").Name("unlockWorkspace")

	r.HandleFunc("/organizations/{organization_name}/workspaces/{workspace_name}/runs", app.listRuns).Methods("GET").Name("listRun")
	r.HandleFunc("/organizations/{organization_name}/workspaces/{workspace_name}/runs/new", app.newRun).Methods("GET").Name("newRun")
	r.HandleFunc("/organizations/{organization_name}/workspaces/{workspace_name}/runs/create", app.createRun).Methods("POST").Name("createRun")
	r.HandleFunc("/organizations/{organization_name}/workspaces/{workspace_name}/runs/{run_id}", app.getRun).Methods("GET").Name("getRun")
	r.HandleFunc("/organizations/{organization_name}/workspaces/{workspace_name}/runs/{run_id}/plan", app.getPlan).Methods("GET").Name("getPlan")
	r.HandleFunc("/organizations/{organization_name}/workspaces/{workspace_name}/runs/{run_id}/apply", app.getApply).Methods("GET").Name("getApply")
	r.HandleFunc("/organizations/{organization_name}/workspaces/{workspace_name}/runs/{run_id}/delete", app.deleteRun).Methods("POST").Name("deleteRun")
}
