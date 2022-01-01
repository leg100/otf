package html

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf"
)

func (app *Application) organizationNewAnchor() anchor {
	return anchor{Name: "new", Link: app.link("organizations", "new")}
}

func (app *Application) organizationShowRoute(organizationID string) string {
	return app.link("organizations", organizationID)
}
func (app *Application) organizationShowAnchor(organizationID string) anchor {
	return anchor{Name: organizationID, Link: app.organizationShowRoute(organizationID)}
}
func (app *Application) organizationShowBreadcrumbs(organization string) []anchor {
	return append(app.organizationListBreadcrumbs(), app.organizationShowAnchor(organization))
}

func (app *Application) organizationListRoute() string { return app.link("organizations") }
func (app *Application) organizationListAnchor() anchor {
	return anchor{Name: "organizations", Link: app.organizationListRoute()}
}
func (app *Application) organizationListBreadcrumbs() []anchor {
	return []anchor{app.organizationListAnchor()}
}

func (app *Application) organizationListHandler(w http.ResponseWriter, r *http.Request) {
	organizations, err := app.OrganizationService().List(otf.OrganizationListOptions{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	opts := []templateDataOption{
		withBreadcrumbs(app.organizationListBreadcrumbs()...),
		withSidebar("organizations", app.organizationNewAnchor()),
	}

	if err := app.render(r, "organizations_list.tmpl", w, organizations, opts...); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (app *Application) organizationsShowHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	organization, err := app.OrganizationService().Get(name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	opts := []templateDataOption{
		withBreadcrumbs(app.organizationShowBreadcrumbs(name)...),
	}

	if err := app.render(r, "organizations_show.tmpl", w, organization, opts...); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (app *Application) organizationsNewHandler(w http.ResponseWriter, r *http.Request) {
	opts := []templateDataOption{
		withBreadcrumbs(app.organizationShowBreadcrumbs(name)...),
	}

	if err := app.render(r, "organizations_show.tmpl", w, organization, opts...); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
