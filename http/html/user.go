package html

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf"
)

func (app *Application) listUsers(w http.ResponseWriter, r *http.Request) {
	opts := otf.UserListOptions{
		OrganizationName: otf.String(mux.Vars(r)["organization_name"]),
	}
	users, err := app.ListUsers(r.Context(), opts)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	app.render("users_list.tmpl", w, r, users)
}
