package authenticator

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf"
)

type htmlApp struct {
	otf.Renderer
}

func (app *htmlApp) AddHandlers(r *mux.Router) {
	// routes that don't require authentication.
	r.HandleFunc("/login", app.loginHandler)
	for _, auth := range app.authenticators {
		r.HandleFunc(auth.RequestPath(), auth.RequestHandler)
		r.HandleFunc(auth.CallbackPath(), auth.responseHandler)
	}
}

func (app *htmlApp) loginHandler(w http.ResponseWriter, r *http.Request) {
	app.Render("login.tmpl", w, r, app.authenticators)
}
