package html

import (
	"html/template"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"golang.org/x/oauth2"
	githubOAuth2 "golang.org/x/oauth2/github"

	"github.com/alexedwards/scs/postgresstore"
	"github.com/alexedwards/scs/v2"
	"github.com/dghubble/gologin/v2"
	"github.com/dghubble/gologin/v2/github"
	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"github.com/leg100/otf"
)

var githubScopes = []string{"user:email", "read:org"}

// Application is the oTF web app.
type Application struct {
	// Sessions manager
	sessions *scs.SessionManager

	// OAuth2 configuration for authorization
	oauth2Config *oauth2.Config

	// HTML template renderer
	renderer

	// Static asset server
	staticServer http.FileSystem

	// oTF service accessors
	otf.Application
}

// NewApplication constructs a new application with the given config
func NewApplication(logger logr.Logger, config Config, services otf.Application, db otf.DB) (*Application, error) {
	if err := config.validate(); err != nil {
		return nil, err
	}

	if config.DevMode {
		logger.Info("enabled developer mode")
	}

	renderer, err := newRenderer(config.DevMode)
	if err != nil {
		return nil, err
	}

	oauth2Config := &oauth2.Config{
		ClientID:     config.GithubClientID,
		ClientSecret: config.GithubClientSecret,
		RedirectURL:  config.GithubRedirectURL,
		Endpoint:     githubOAuth2.Endpoint,
		Scopes:       githubScopes,
	}

	sessions := scs.New()
	sessions.Store = postgresstore.New(db.Handle().DB)

	app := &Application{
		Application:  services,
		sessions:     sessions,
		oauth2Config: oauth2Config,
		renderer:     renderer,
		staticServer: newStaticServer(config.DevMode),
	}

	return app, nil
}

// AddRoutes adds application routes and middleware to an HTTP multiplexer.
func (app *Application) AddRoutes(router *mux.Router) {
	// Static assets (JS, CSS, etc).
	router.PathPrefix("/static/").Handler(http.FileServer(app.staticServer)).Methods("GET")

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
	router.Use(newStateHandler(gologin.DebugOnlyCookieConfig))

	router.Handle("/github/login", LoginHandler(app.oauth2Config, nil))
	router.Handle(githubCallbackPath, github.CallbackHandler(app.oauth2Config, http.HandlerFunc(app.githubLogin), nil))

}

// authRoutes adds routes that require authentication.
func (app *Application) authRoutes(router *mux.Router) {
	router.Use(app.requireAuthentication)

	router.HandleFunc("/profile", app.profileHandler).Methods("GET")
	router.HandleFunc("/sessions", app.sessionsHandler).Methods("GET")
	router.HandleFunc("/sessions/revoke", app.revokeSessionHandler).Methods("POST")
}

// render wraps calls to the template renderer, adding common data to the
// template
func (app *Application) render(r *http.Request, name string, w io.Writer, content interface{}, opts ...templateDataOption) error {
	data := templateData{
		Title:       strings.Title(filenameWithoutExtension(name)),
		Content:     content,
		CurrentUser: app.currentUser(r),
	}

	for _, o := range opts {
		o(&data)
	}

	// Get flash msg
	if msg := app.sessions.PopString(r.Context(), otf.FlashSessionKey); msg != "" {
		data.Flash = template.HTML(msg)
	}

	return app.renderTemplate(name, w, data)
}

func filenameWithoutExtension(fname string) string {
	return strings.TrimSuffix(fname, filepath.Ext(fname))
}
