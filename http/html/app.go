package html

import (
	"net/http"

	"github.com/go-logr/logr"
	otfhttp "github.com/leg100/otf/http"
)

// Application is the otf web app.
type Application struct {
	staticServer http.FileSystem // Static asset server
	pathPrefix   string          // path prefix for all URLs

	logr.Logger // logger for logging messages
}

// ApplicationOptions are options for configuring the web app
type ApplicationOptions struct {
	*otfhttp.ServerConfig
	*otfhttp.Router
}

// AddRoutes adds routes for the html web app.
func AddRoutes(logger logr.Logger, opts ApplicationOptions) error {
	app := &Application{
		staticServer: newStaticServer(opts.DevMode),
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
		r.Use(setOrganization)

		// Session routes
		app.sessionService.AddHTMLHandlers(r.Router)

		// User routes
		app.userService.AddHTMLHandlers(r.Router)

		// Module routes
		app.moduleService.AddHTMLHandlers(r.Router)

		// VCS provider routes
		app.vcsProviderService.AddHTMLHandlers(r.Router)

		// Variables routes
		app.variableService.AddHTMLHandlers(r.Router)
		// Run routes
		app.runService.AddHTMLHandlers(r.Router)
		// Watch routes
		app.watchService.AddHTMLHandlers(r.Router)
	})
}
