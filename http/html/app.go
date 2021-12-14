package html

import (
	"html/template"
	"io"
	"net/http"

	"golang.org/x/oauth2"
	githubOAuth2 "golang.org/x/oauth2/github"

	"github.com/alexedwards/scs/v2"
	"github.com/dghubble/gologin/v2"
	"github.com/dghubble/gologin/v2/github"
	"github.com/gorilla/mux"
)

var githubScopes = []string{"user:email", "read:org"}

// application is the oTF web app.
type application struct {
	// Sessions manager
	sessions *scs.SessionManager

	// OAuth2 configuration for authorization
	oauth2Config *oauth2.Config

	// HTML template renderer
	renderer

	// Static asset server
	staticServer http.FileSystem
}

// NewApplication constructs a new application with the given config
func NewApplication(config *Config) (*application, error) {
	if err := config.validate(); err != nil {
		return nil, err
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

	app := &application{
		sessions:     scs.New(),
		oauth2Config: oauth2Config,
		renderer:     renderer,
		staticServer: newStaticServer(config.DevMode),
	}

	return app, nil
}

// AddRoutes adds application routes and middleware to an HTTP multiplexer.
func (app *application) AddRoutes(router *mux.Router) {
	router = router.NewRoute().Subrouter()

	// Static assets (JS, CSS, etc). Ensure this is before enabling sessions
	// middleware to avoid setting cookies unnecessarily.
	router.PathPrefix("/static/").Handler(http.FileServer(app.staticServer)).Methods("GET")

	// Enable sessions middleware
	router.Use(app.sessions.LoadAndSave)

	stateConfig := gologin.DebugOnlyCookieConfig
	router.Handle("/github/login", github.StateHandler(stateConfig, github.LoginHandler(app.oauth2Config, nil)))
	router.Handle(githubCallbackPath, github.StateHandler(stateConfig, github.CallbackHandler(app.oauth2Config, app.issueSession(), nil)))

	router.HandleFunc("/login", app.loginHandler).Methods("GET")
	router.HandleFunc("/logout", app.logoutHandler).Methods("POST")

	router = router.NewRoute().Subrouter()
	router.Use(app.requireAuthentication)

	router.HandleFunc("/profile", app.profileHandler).Methods("GET")
}

// render wraps calls to the template renderer, adding common data to the
// template
func (app *application) render(r *http.Request, name string, w io.Writer, content interface{}) error {
	data := templateData{
		Content:         content,
		IsAuthenticated: app.isAuthenticated(r),
	}

	// Get flash msg
	if msg := app.sessions.PopString(r.Context(), sessionFlashKey); msg != "" {
		data.Flash = template.HTML(msg)
	}

	return app.renderTemplate(name, w, data)
}
