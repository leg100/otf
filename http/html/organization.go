package html

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/leg100/otf/http/decode"
)

func (app *Application) newOrganization(w http.ResponseWriter, r *http.Request) {
	app.render("organization_new.tmpl", w, r, nil)
}

func (app *Application) createOrganization(w http.ResponseWriter, r *http.Request) {
	var opts otf.OrganizationCreateOptions
	if err := decode.Form(&opts, r); err != nil {
		writeError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	org, err := app.CreateOrganization(r.Context(), opts)
	if err == otf.ErrResourceAlreadyExists {
		flashError(w, "organization already exists: "+*opts.Name)
		http.Redirect(w, r, newOrganizationPath(), http.StatusFound)
		return
	}
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	flashSuccess(w, "created organization: "+org.Name())
	http.Redirect(w, r, organizationPath(org.Name()), http.StatusFound)
}

func (app *Application) listOrganizations(w http.ResponseWriter, r *http.Request) {
	var opts otf.OrganizationListOptions
	if err := decode.Query(&opts, r.URL.Query()); err != nil {
		writeError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	organizations, err := app.ListOrganizations(r.Context(), opts)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	app.render("organization_list.tmpl", w, r, struct {
		*otf.OrganizationList
		otf.OrganizationListOptions
	}{
		OrganizationList:        organizations,
		OrganizationListOptions: opts,
	})
}

func (app *Application) getOrganization(w http.ResponseWriter, r *http.Request) {
	org, err := app.GetOrganization(r.Context(), mux.Vars(r)["organization_name"])
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	app.render("organization_get.tmpl", w, r, org)
}

func (app *Application) editOrganization(w http.ResponseWriter, r *http.Request) {
	organization, err := app.GetOrganization(r.Context(), mux.Vars(r)["organization_name"])
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	app.render("organization_edit.tmpl", w, r, organization)
}

func (app *Application) updateOrganization(w http.ResponseWriter, r *http.Request) {
	var opts otf.OrganizationUpdateOptions
	if err := decode.Form(&opts, r); err != nil {
		writeError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	org, err := app.UpdateOrganization(r.Context(), mux.Vars(r)["organization_name"], &opts)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	flashSuccess(w, "updated organization")
	http.Redirect(w, r, editOrganizationPath(org.Name()), http.StatusFound)
}

func (app *Application) deleteOrganization(w http.ResponseWriter, r *http.Request) {
	organizationName := mux.Vars(r)["organization_name"]
	err := app.DeleteOrganization(r.Context(), organizationName)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	flashSuccess(w, "deleted organization: "+organizationName)
	http.Redirect(w, r, organizationsPath(), http.StatusFound)
}

func (app *Application) listOrganizationPermissions(w http.ResponseWriter, r *http.Request) {
	org, err := app.GetOrganization(r.Context(), mux.Vars(r)["organization_name"])
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	app.render("organization_get.tmpl", w, r, org)
}
