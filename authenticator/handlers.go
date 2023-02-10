package authenticator

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf"
)

type handlers struct {
	otf.Renderer

	authenticators []*Authenticator
}

func (app *handlers) AddHandlers(r *mux.Router) {
	// routes that don't require authentication.
	r.HandleFunc("/login", app.loginHandler)
	for _, auth := range app.authenticators {
		r.HandleFunc(auth.RequestPath(), auth.RequestHandler)
		r.HandleFunc(auth.CallbackPath(), auth.responseHandler)
	}
}

func (app *handlers) loginHandler(w http.ResponseWriter, r *http.Request) {
	app.Render("login.tmpl", w, r, app.authenticators)
}
