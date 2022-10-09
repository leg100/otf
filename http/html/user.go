package html

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf"
)

type UserList struct {
	Items []*otf.User
	opts  otf.UserListOptions
}

func (l UserList) OrganizationName() string {
	if l.opts.OrganizationName == nil {
		return ""
	}
	return *l.opts.OrganizationName
}

func (l UserList) TeamName() string {
	if l.opts.TeamName == nil {
		return ""
	}
	return *l.opts.TeamName
}

func (app *Application) listUsers(w http.ResponseWriter, r *http.Request) {
	opts := otf.UserListOptions{
		OrganizationName: otf.String(mux.Vars(r)["organization_name"]),
	}
	users, err := app.ListUsers(r.Context(), opts)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	app.render("users_list.tmpl", w, r, UserList{users, opts})
}
