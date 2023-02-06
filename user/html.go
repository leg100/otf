package user

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf"
)

func (app *Application) listUsers(w http.ResponseWriter, r *http.Request) {
	opts := otf.UserListOptions{
		Organization: otf.String(mux.Vars(r)["organization_name"]),
	}
	users, err := app.ListUsers(r.Context(), opts)
	if err != nil {
		Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	app.Render("users_list.tmpl", w, r, users)
}