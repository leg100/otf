package html

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/leg100/otf/http/decode"
)

// organizationRequest provides metadata about a request for a organization
type organizationRequest struct {
	r *http.Request
}

func (w organizationRequest) OrganizationName() string {
	return param(w.r, "organization_name")
}

func (app *Application) listOrganizations(w http.ResponseWriter, r *http.Request) {
	var opts otf.OrganizationListOptions
	if err := decode.Query(&opts, r.URL.Query()); err != nil {
		writeError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	organizations, err := app.OrganizationService().List(r.Context(), opts)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	app.render("organization_list.tmpl", w, r, organizations)
}

func (app *Application) newOrganization(w http.ResponseWriter, r *http.Request) {
	app.render("organization_new.tmpl", w, r, nil)
}

func (app *Application) createOrganization(w http.ResponseWriter, r *http.Request) {
	var opts otf.OrganizationCreateOptions
	if err := decode.Form(&opts, r); err != nil {
		writeError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	org, err := app.OrganizationService().Create(r.Context(), opts)
	if err == otf.ErrResourcesAlreadyExists {
		flashError(w, "organization already exists: "+*opts.Name)
		http.Redirect(w, r, newOrganizationPath(), http.StatusFound)
		return
	}
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	flashSuccess(w, "created organization: "+org.Name())
	http.Redirect(w, r, getOrganizationPath(org), http.StatusFound)
}

// Get lists the workspaces for the org.
func (app *Application) getOrganization(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, listWorkspacePath(organizationRequest{r}), http.StatusFound)
}

func (app *Application) getOrganizationOverview(w http.ResponseWriter, r *http.Request) {
	org, err := app.OrganizationService().Get(r.Context(), mux.Vars(r)["organization_name"])
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	app.render("organization_get.tmpl", w, r, org)
}

func (app *Application) editOrganization(w http.ResponseWriter, r *http.Request) {
	organization, err := app.OrganizationService().Get(r.Context(), mux.Vars(r)["organization_name"])
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
	org, err := app.OrganizationService().Update(r.Context(), mux.Vars(r)["organization_name"], &opts)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	flashSuccess(w, "updated organization")
	http.Redirect(w, r, editOrganizationPath(org), http.StatusFound)
}

func (app *Application) deleteOrganization(w http.ResponseWriter, r *http.Request) {
	organizationName := mux.Vars(r)["organization_name"]
	err := app.OrganizationService().Delete(r.Context(), organizationName)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	flashSuccess(w, "deleted organization: "+organizationName)
	http.Redirect(w, r, listOrganizationPath(), http.StatusFound)
}
