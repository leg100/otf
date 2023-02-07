package user

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/leg100/otf/http/decode"
	"github.com/leg100/otf/http/html"
)

type htmlApp struct {
	otf.Renderer

	app service
}

func (app *htmlApp) AddHTMLHandlers(r *mux.Router) {
	r.HandleFunc("/organizations/{organization_name}/users", app.listUsers).Methods("GET")
}

func (app *htmlApp) listUsers(w http.ResponseWriter, r *http.Request) {
	organization, err := decode.Param("organization_name", r)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	users, err := app.ListUsers(r.Context(), UserListOptions{
		Organization: otf.String(organization),
	})
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	app.Render("users_list.tmpl", w, r, users)
}
