package html

import (
	"io"
	"net/http"

	"golang.org/x/oauth2"
	githubOAuth2 "golang.org/x/oauth2/github"

	"github.com/alexedwards/scs/v2"
	"github.com/dghubble/gologin/v2"
	"github.com/dghubble/gologin/v2/github"
	"github.com/gorilla/mux"
	"github.com/leg100/otf/http/html/assets"
)

const (
	sessionUserKey  = "githubID"
	sessionUsername = "githubUsername"
)

var (
	embeddedAssetServer assets.Server

	sessions = scs.New()

	homeHtml = `<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Github Example</title>
    <style>
      a.button {
        text-decoration: none;
      }
    </style>
  </head>

  <body>
    <a href="/github/login" class="button">Login with Github</a>
  </body>
</html>
`
)

type Config struct {
	GithubClientID     string
	GithubClientSecret string
}

// Load embedded templates at startup
func init() {
	server, err := assets.NewEmbeddedServer()
	if err != nil {
		panic("unable to load embedded assets: " + err.Error())
	}

	embeddedAssetServer = server
}

// New configures and adds the app to the HTTP mux.
func New(router *mux.Router, config *Config) {
	// 1. Register LoginHandler and CallbackHandler
	oauth2Config := &oauth2.Config{
		ClientID:     config.GithubClientID,
		ClientSecret: config.GithubClientSecret,
		RedirectURL:  "https://localhost:8080/github/callback",
		Endpoint:     githubOAuth2.Endpoint,
		Scopes:       []string{"user:email", "read:org"},
	}

	router = router.NewRoute().Subrouter()

	router.Use(sessions.LoadAndSave)

	router.HandleFunc("/profile", profileHandler).Methods("GET")

	stateConfig := gologin.DebugOnlyCookieConfig
	router.Handle("/github/login", github.StateHandler(stateConfig, github.LoginHandler(oauth2Config, nil)))
	router.Handle("/github/callback", github.StateHandler(stateConfig, github.CallbackHandler(oauth2Config, issueSession(), nil)))

	router.HandleFunc("/logout", logoutHandler).Methods("POST")
}

// issueSession issues a cookie session after successful Github login
func issueSession() http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		githubUser, err := github.UserFromContext(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		sessions.Put(r.Context(), sessionUserKey, *githubUser.ID)
		sessions.Put(r.Context(), sessionUsername, *githubUser.Login)

		http.Redirect(w, r, "/profile", http.StatusFound)
	}
	return http.HandlerFunc(fn)
}

func profileHandler(w http.ResponseWriter, r *http.Request) {
	username := sessions.GetString(r.Context(), sessionUsername)
	if username == "" {
		io.WriteString(w, homeHtml)
		return
	}

	io.WriteString(w, "You are logged in as: "+username)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	if err := sessions.Destroy(r.Context()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}
