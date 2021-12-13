package html

import (
	"context"
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

const (
	sessionUserKey  = "githubID"
	sessionUsername = "githubUsername"
	sessionFlashKey = "flash"
)

type Application struct {
	sessions *scs.SessionManager

	oauth2Config *oauth2.Config

	renderer
}

type Config struct {
	GithubClientID     string
	GithubClientSecret string
	DevMode            bool
}

// NewApplication constructs a new application with the given config
func NewApplication(config *Config) (*Application, error) {
	// 1. Register LoginHandler and CallbackHandler
	oauth2Config := &oauth2.Config{
		ClientID:     config.GithubClientID,
		ClientSecret: config.GithubClientSecret,
		RedirectURL:  "https://localhost:8080/github/callback",
		Endpoint:     githubOAuth2.Endpoint,
		Scopes:       []string{"user:email", "read:org"},
	}

	renderer, err := newRenderer(config.DevMode)
	if err != nil {
		return nil, err
	}

	app := &Application{
		sessions:     scs.New(),
		oauth2Config: oauth2Config,
		renderer:     renderer,
	}

	return app, nil
}

// AddRoutes adds application routes and middleware to an HTTP multiplexer.
func (app *Application) AddRoutes(router *mux.Router) {
	router = router.NewRoute().Subrouter()

	// Enable sessions middleware
	router.Use(app.sessions.LoadAndSave)

	router.HandleFunc("/profile", app.profileHandler).Methods("GET")

	stateConfig := gologin.DebugOnlyCookieConfig
	router.Handle("/github/login", github.StateHandler(stateConfig, github.LoginHandler(app.oauth2Config, nil)))
	router.Handle("/github/callback", github.StateHandler(stateConfig, github.CallbackHandler(app.oauth2Config, app.issueSession(), nil)))

	router.HandleFunc("/login", app.loginHandler).Methods("GET")
	router.HandleFunc("/logout", app.logoutHandler).Methods("POST")
}

// render wraps calls to the template renderer, adding common data such as a
// flash message.
func (app *Application) render(ctx context.Context, name string, w io.Writer, content interface{}) error {
	data := TemplateData{
		Content: content,
	}

	// Get flash msg
	if msg := app.sessions.GetString(ctx, sessionFlashKey); msg != "" {
		data.Flash = template.HTML(msg)
	}

	return app.renderTemplate(name, w, data)
}

// issueSession issues a cookie session after successful Github login
func (app *Application) issueSession() http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		githubUser, err := github.UserFromContext(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		app.sessions.Put(r.Context(), sessionUserKey, *githubUser.ID)
		app.sessions.Put(r.Context(), sessionUsername, *githubUser.Login)

		http.Redirect(w, r, "/profile", http.StatusFound)
	}
	return http.HandlerFunc(fn)
}

func (app *Application) profileHandler(w http.ResponseWriter, r *http.Request) {
	username := app.sessions.GetString(r.Context(), sessionUsername)
	if username == "" {
		io.WriteString(w, "")
		return
	}

	io.WriteString(w, "You are logged in as: "+username)
}

func (app *Application) loginHandler(w http.ResponseWriter, r *http.Request) {
	if err := app.render(r.Context(), "login.tmpl", w, nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (app *Application) logoutHandler(w http.ResponseWriter, r *http.Request) {
	if err := app.sessions.Destroy(r.Context()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}
