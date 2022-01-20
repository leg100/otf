package html

import (
	"net/http"

	"github.com/alexedwards/scs/postgresstore"
	"github.com/alexedwards/scs/v2"
	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"github.com/leg100/otf"
)

const DefaultPathPrefix = "/"

var (
	getProfileRoute *mux.Route
)

// Application is the oTF web app.
type Application struct {
	// Sessions manager
	sessions *scs.SessionManager

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

// NewApplication constructs a new application with the given config
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

	sessions := scs.New()
	sessions.Store = postgresstore.New(db.Handle().DB)

	app := &Application{
		Application:  services,
		sessions:     sessions,
		oauth:        oauthApp,
		renderer:     renderer,
		staticServer: newStaticServer(config.DevMode),
		pathPrefix:   DefaultPathPrefix,
		templateDataFactory: &templateDataFactory{
			sessions: sessions,
			router:   muxrouter,
		},
		router: &router{Router: muxrouter},
	}

	app.addRoutes(muxrouter)

	return nil
}

// AddRoutes adds application routes and middleware to an HTTP multiplexer.
func (app *Application) addRoutes(router *mux.Router) {
	// Static assets (JS, CSS, etc).
	router.PathPrefix("/static/").Handler(http.FileServer(app.staticServer)).Methods("GET")

	// Redirect paths with a trailing slash to path without, e.g. /runs/ ->
	// /runs. Uses an HTTP301.
	router.StrictSlash(true)

	app.sessionRoutes(router.NewRoute().Subrouter())
}

// sessionRoutes adds routes for which a session is maintained.
func (app *Application) sessionRoutes(router *mux.Router) {
	// Enable sessions middleware
	router.Use(app.sessions.LoadAndSave)

	app.nonAuthRoutes(router.NewRoute().Subrouter())
	app.authRoutes(router.NewRoute().Subrouter())
}

// nonAuthRoutes adds routes that don't require authentication.
func (app *Application) nonAuthRoutes(router *mux.Router) {
	app.githubRoutes(router.NewRoute().Subrouter())

	router.HandleFunc("/login", app.loginHandler).Methods("GET").Name("login")
	router.HandleFunc("/logout", app.logoutHandler).Methods("POST").Name("logout")
}

func (app *Application) githubRoutes(router *mux.Router) {
	router.HandleFunc("/github/login", app.oauth.requestHandler)
	router.HandleFunc(githubCallbackPath, app.githubLogin)
}

// authRoutes adds routes that require authentication.
func (app *Application) authRoutes(router *mux.Router) {
	router.Use(app.requireAuthentication)

	router.HandleFunc("/me", app.meHandler).Methods("GET").Name("getMe")
	getProfileRoute = router.HandleFunc("/me/profile", app.profileHandler).Methods("GET").Name("getProfile")
	router.HandleFunc("/me/sessions", app.sessionsHandler).Methods("GET").Name("listSession")
	router.HandleFunc("/me/sessions/revoke", app.revokeSessionHandler).Methods("POST").Name("revokeSession")

	// TODO: replace sessions handler with token handler when one exists
	router.HandleFunc("/tokens", app.sessionsHandler).Methods("GET").Name("listToken")

	(&OrganizationController{
		OrganizationService: app.OrganizationService(),
		templateDataFactory: app.templateDataFactory,
		renderer:            app.renderer,
		router:              app.router,
		sessions:            &sessions{app.sessions},
	}).addRoutes(router.PathPrefix("/organizations").Subrouter())

	(&WorkspaceController{
		WorkspaceService:    app.WorkspaceService(),
		templateDataFactory: app.templateDataFactory,
		renderer:            app.renderer,
		router:              app.router,
		sessions:            &sessions{app.sessions},
	}).addRoutes(router.PathPrefix("/organizations/{organization_name}/workspaces").Subrouter())

	(&RunController{
		RunService:          app.RunService(),
		PlanService:         app.PlanService(),
		templateDataFactory: app.templateDataFactory,
		renderer:            app.renderer,
		router:              app.router,
		sessions:            &sessions{app.sessions},
	}).addRoutes(router.PathPrefix("/organizations/{organization_name}/workspaces/{workspace_name}/runs").Subrouter())
}
