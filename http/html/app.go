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
	*templateDataFactory
	// wrapper around mux router
	*router
}

// AddRoutes adds routes for the html web app.
func AddRoutes(logger logr.Logger, config Config, services otf.Application, db otf.DB, muxrouter *mux.Router) error {
	if config.DevMode {
		logger.Info("enabled developer mode")
	}

	renderer, err := newRenderer(config.DevMode)
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
		renderer:     renderer,
		staticServer: newStaticServer(config.DevMode),
		pathPrefix:   DefaultPathPrefix,
		templateDataFactory: &templateDataFactory{
			router: muxrouter,
		},
		router: &router{Router: muxrouter},
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
func (app *Application) authRoutes(router *mux.Router) {
	router.Use(app.authenticateUser)
	router.Use(app.setCurrentOrganization)

	router.HandleFunc("/logout", app.logoutHandler).Methods("POST").Name("logout")
	router.HandleFunc("/profile", app.profileHandler).Methods("GET").Name("getProfile")
	router.HandleFunc("/profile/sessions", app.sessionsHandler).Methods("GET").Name("listSession")
	router.HandleFunc("/profile/sessions/revoke", app.revokeSessionHandler).Methods("POST").Name("revokeSession")

	router.HandleFunc("/profile/tokens", app.tokensHandler).Methods("GET").Name("listToken")
	router.HandleFunc("/profile/tokens/delete", app.deleteTokenHandler).Methods("POST").Name("deleteToken")
	router.HandleFunc("/profile/tokens/new", app.newTokenHandler).Methods("GET").Name("newToken")
	router.HandleFunc("/profile/tokens/create", app.createTokenHandler).Methods("POST").Name("createToken")

	(&OrganizationController{
		OrganizationService: app.OrganizationService(),
		templateDataFactory: app.templateDataFactory,
		renderer:            app.renderer,
		router:              app.router,
	}).addRoutes(router.PathPrefix("/organizations").Subrouter())

	(&WorkspaceController{
		WorkspaceService:    app.WorkspaceService(),
		templateDataFactory: app.templateDataFactory,
		renderer:            app.renderer,
		router:              app.router,
	}).addRoutes(router.PathPrefix("/organizations/{organization_name}/workspaces").Subrouter())

	(&RunController{
		RunService:          app.RunService(),
		PlanService:         app.PlanService(),
		ApplyService:        app.ApplyService(),
		WorkspaceService:    app.WorkspaceService(),
		templateDataFactory: app.templateDataFactory,
		renderer:            app.renderer,
		router:              app.router,
	}).addRoutes(router.PathPrefix("/organizations/{organization_name}/workspaces/{workspace_name}/runs").Subrouter())
}
