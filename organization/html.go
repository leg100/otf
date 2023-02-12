package organization

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/leg100/otf/http/decode"
	"github.com/leg100/otf/http/html"
	"github.com/leg100/otf/http/paths"
)

// webApp is the web application for organizations
type webApp struct {
	otf.Renderer // renders templates

	app appService // provide access to org service
}

func (a *webApp) AddHTMLHandlers(r *mux.Router) {
	r.HandleFunc("/organizations", a.listOrganizations)
	r.HandleFunc("/organizations/new", a.newOrganization)
	r.HandleFunc("/organizations/create", a.createOrganization)
	r.HandleFunc("/organizations/{organization_name}", a.getOrganization)
	r.HandleFunc("/organizations/{organization_name}/edit", a.editOrganization)
	r.HandleFunc("/organizations/{organization_name}/update", a.updateOrganization)
	r.HandleFunc("/organizations/{organization_name}/delete", a.deleteOrganization)
	r.HandleFunc("/organizations/{organization_name}/permissions", a.listOrganizationPermissions)
}

func (a *webApp) newOrganization(w http.ResponseWriter, r *http.Request) {
	a.Render("organization_new.tmpl", w, r, nil)
}

func (a *webApp) createOrganization(w http.ResponseWriter, r *http.Request) {
	var opts OrganizationCreateOptions
	if err := decode.Form(&opts, r); err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	org, err := a.app.create(r.Context(), opts)
	if err == otf.ErrResourceAlreadyExists {
		html.FlashError(w, "organization already exists: "+*opts.Name)
		http.Redirect(w, r, paths.NewOrganization(), http.StatusFound)
		return
	}
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	html.FlashSuccess(w, "created organization: "+org.Name())
	http.Redirect(w, r, paths.Organization(org.Name()), http.StatusFound)
}

func (a *webApp) listOrganizations(w http.ResponseWriter, r *http.Request) {
	var opts OrganizationListOptions
	if err := decode.Query(&opts, r.URL.Query()); err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	organizations, err := a.app.ListOrganizations(r.Context(), opts)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	a.Render("organization_list.tmpl", w, r, struct {
		*organizationList
		OrganizationListOptions
	}{
		organizationList:        organizations,
		OrganizationListOptions: opts,
	})
}

func (a *webApp) getOrganization(w http.ResponseWriter, r *http.Request) {
	org, err := a.app.GetOrganization(r.Context(), mux.Vars(r)["organization_name"])
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	a.Render("organization_get.tmpl", w, r, org)
}

func (a *webApp) editOrganization(w http.ResponseWriter, r *http.Request) {
	organization, err := a.app.GetOrganization(r.Context(), mux.Vars(r)["organization_name"])
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	a.Render("organization_edit.tmpl", w, r, organization)
}

func (a *webApp) updateOrganization(w http.ResponseWriter, r *http.Request) {
	var opts OrganizationUpdateOptions
	if err := decode.Form(&opts, r); err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	org, err := a.app.UpdateOrganization(r.Context(), mux.Vars(r)["organization_name"], &opts)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	html.FlashSuccess(w, "updated organization")
	http.Redirect(w, r, paths.EditOrganization(org.Name()), http.StatusFound)
}

func (a *webApp) deleteOrganization(w http.ResponseWriter, r *http.Request) {
	organizationName := mux.Vars(r)["organization_name"]
	err := a.app.DeleteOrganization(r.Context(), organizationName)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	html.FlashSuccess(w, "deleted organization: "+organizationName)
	http.Redirect(w, r, paths.Organizations(), http.StatusFound)
}

func (a *webApp) listOrganizationPermissions(w http.ResponseWriter, r *http.Request) {
	org, err := a.app.GetOrganization(r.Context(), mux.Vars(r)["organization_name"])
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	a.Render("organization_get.tmpl", w, r, org)
}
