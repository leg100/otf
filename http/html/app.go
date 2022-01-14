package html

import (
	"net/http"
	"path/filepath"
	"strings"

	"github.com/alexedwards/scs/postgresstore"
	"github.com/alexedwards/scs/v2"
	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"github.com/leg100/otf"
)

const DefaultPathPrefix = "/"

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
}

// NewApplication constructs a new application with the given config
func AddRoutes(logger logr.Logger, config Config, services otf.Application, db otf.DB, router *mux.Router) error {
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
			router:   router,
		},
	}

	app.addRoutes(router)

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

	router.HandleFunc("/login", app.loginHandler).Methods("GET")
	router.HandleFunc("/logout", app.logoutHandler).Methods("POST")
}

func (app *Application) githubRoutes(router *mux.Router) {
	router.HandleFunc("/github/login", app.oauth.requestHandler)
	router.HandleFunc(githubCallbackPath, app.githubLogin)
}

// authRoutes adds routes that require authentication.
func (app *Application) authRoutes(router *mux.Router) {
	router.Use(app.requireAuthentication)

	router.HandleFunc("/profile", app.profileHandler).Methods("GET").Name("getProfile")
	router.HandleFunc("/sessions", app.sessionsHandler).Methods("GET").Name("listSession")
	router.HandleFunc("/sessions/revoke", app.revokeSessionHandler).Methods("POST").Name("revokeSession")

	// TODO: replace sessions handler with token handler when one exists
	router.HandleFunc("/tokens", app.sessionsHandler).Methods("GET").Name("listToken")

	(&OrganizationController{
		OrganizationService: app.OrganizationService(),
		templateDataFactory: app.templateDataFactory,
		renderer:            app.renderer,
	}).addRoutes(router.PathPrefix("/organizations").Subrouter())

	(&WorkspaceController{
		WorkspaceService:    app.WorkspaceService(),
		templateDataFactory: app.templateDataFactory,
		renderer:            app.renderer,
	}).addRoutes(router.PathPrefix("/organizations/{organization_name}/workspaces").Subrouter())

	(&RunController{
		RunService:          app.RunService(),
		templateDataFactory: app.templateDataFactory,
		renderer:            app.renderer,
	}).addRoutes(router.PathPrefix("/organizations/{organization_name}/workspaces/{workspace_name}/runs").Subrouter())
}

// link produces a relative link for the site
func (app *Application) link(path ...string) string {
	return filepath.Join(append([]string{app.pathPrefix}, path...)...)
}

func filenameWithoutExtension(fname string) string {
	return strings.TrimSuffix(fname, filepath.Ext(fname))
}
