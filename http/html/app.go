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
	staticServer       http.FileSystem  // Static asset server
	pathPrefix         string           // path prefix for all URLs
	authenticators     []*Authenticator // enabled authenticators
	siteToken          string           // site admin's authentication token
	secret             string           // secret for webhook signatures
	variableService    otf.VariableService
	vcsProviderService otf.VCSProviderService

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
	otf.VCSProviderService
}

// AddRoutes adds routes for the html web app.
func AddRoutes(logger logr.Logger, opts ApplicationOptions) error {
	logger = logger.WithValues("component", "html")
	if opts.DevMode {
		logger.Info("enabled developer mode")
	}

	viewEngine, err := NewViewEngine(opts.DevMode)
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
		Application:        opts.Application,
		variableService:    opts.VariableService,
		vcsProviderService: opts.VCSProviderService,
		staticServer:       newStaticServer(opts.DevMode),
		pathPrefix:         DefaultPathPrefix,
		viewEngine:         viewEngine,
		Logger:             logger,
		Server:             sseServer,
		siteToken:          opts.SiteToken,
		secret:             opts.Secret,
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

	// TODO: add authenticator handlers here

	// TODO: add session handlers for admin login here

	// routes that require authentication.
	r.Sub(func(r *otfhttp.Router) {
		r.Use((&authMiddleware{app}).authenticate)
		r.Use(setOrganization)

		r.PST("/logout", app.logoutHandler)
		r.GET("/profile", app.profileHandler)

		// Session routes
		app.sessionService.AddHTMLHandlers(r.Router)

		// User routes
		app.userService.AddHTMLHandlers(r.Router)

		// Module routes
		app.moduleService.AddHTMLHandlers(r.Router)

		// VCS provider routes
		app.vcsProviderService.AddHTMLHandlers(r.Router)

		r.GET("/organizations/{organization_name}/permissions", app.listOrganizationPermissions)

		// Variables routes
		app.variableService.AddHTMLHandlers(r.Router)
		// Run routes
		app.runService.AddHTMLHandlers(r.Router)
		// Watch routes
		app.watchService.AddHTMLHandlers(r.Router)
	})
}
