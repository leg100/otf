package organization

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/leg100/otf/http/decode"
	"github.com/leg100/otf/http/html"
	"github.com/leg100/otf/http/html/paths"
)

// web is the web application for organizations
type web struct {
	otf.Renderer

	app application
}

func (a *web) AddHandlers(r *mux.Router) {
	r.HandleFunc("/organizations", a.list)
	r.HandleFunc("/organizations/new", a.new)
	r.HandleFunc("/organizations/create", a.createOrganization)
	r.HandleFunc("/organizations/{name}", a.getOrganization)
	r.HandleFunc("/organizations/{name}/edit", a.editOrganization)
	r.HandleFunc("/organizations/{name}/update", a.update)
	r.HandleFunc("/organizations/{name}/delete", a.delete)
	r.HandleFunc("/organizations/{name}/permissions", a.listPermissions)
}

func (a *web) new(w http.ResponseWriter, r *http.Request) {
	a.Render("organization_new.tmpl", w, r, nil)
}

func (a *web) createOrganization(w http.ResponseWriter, r *http.Request) {
	var opts otf.OrganizationCreateOptions
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

func (a *web) list(w http.ResponseWriter, r *http.Request) {
	var opts listOptions
	if err := decode.Query(&opts, r.URL.Query()); err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	organizations, err := a.app.list(r.Context(), opts)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	a.Render("organization_list.tmpl", w, r, struct {
		*organizationList
		listOptions
	}{
		organizationList: organizations,
		listOptions:      opts,
	})
}

func (a *web) getOrganization(w http.ResponseWriter, r *http.Request) {
	name, err := decode.Param("name", r)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	org, err := a.app.get(r.Context(), name)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	a.Render("organization_get.tmpl", w, r, org)
}

func (a *web) editOrganization(w http.ResponseWriter, r *http.Request) {
	name, err := decode.Param("name", r)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	organization, err := a.app.get(r.Context(), name)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	a.Render("organization_edit.tmpl", w, r, organization)
}

func (a *web) update(w http.ResponseWriter, r *http.Request) {
	params := struct {
		Options UpdateOptions
		Name    string `schema:"name,required"`
	}{}
	if err := decode.All(&params, r); err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	org, err := a.app.update(r.Context(), params.Name, &params.Options)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	html.FlashSuccess(w, "updated organization")
	http.Redirect(w, r, paths.EditOrganization(org.Name()), http.StatusFound)
}

func (a *web) delete(w http.ResponseWriter, r *http.Request) {
	organization, err := decode.Param("name", r)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	err = a.app.delete(r.Context(), organization)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	html.FlashSuccess(w, "deleted organization: "+organization)
	http.Redirect(w, r, paths.Organizations(), http.StatusFound)
}

func (a *web) listPermissions(w http.ResponseWriter, r *http.Request) {
	name, err := decode.Param("name", r)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	org, err := a.app.get(r.Context(), name)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	a.Render("organization_get.tmpl", w, r, org)
}
