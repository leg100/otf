package html

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/leg100/otf/http/decode"
)

func (c *Application) listOrganizations(w http.ResponseWriter, r *http.Request) {
	var opts otf.OrganizationListOptions
	if err := decode.Query(&opts, r.URL.Query()); err != nil {
		writeError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	organizations, err := c.OrganizationService().List(r.Context(), opts)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	c.render("organization_list.tmpl", w, r, organizations)
}

func (c *Application) newOrganization(w http.ResponseWriter, r *http.Request) {
	c.render("organization_new.tmpl", w, r, nil)
}

func (c *Application) createOrganization(w http.ResponseWriter, r *http.Request) {
	var opts otf.OrganizationCreateOptions
	if err := decode.Form(&opts, r); err != nil {
		writeError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	organization, err := c.OrganizationService().Create(r.Context(), opts)
	if err == otf.ErrResourcesAlreadyExists {
		flashError(w, "organization already exists: "+*opts.Name)
		http.Redirect(w, r, c.route("newOrganization"), http.StatusFound)
		return
	}
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	flashSuccess(w, "created organization: "+organization.Name())
	http.Redirect(w, r, c.relative(r, "getOrganization", "organization_name", *opts.Name), http.StatusFound)
}

// Get lists the workspaces for the org.
func (c *Application) getOrganization(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, c.relative(r, "listWorkspace"), http.StatusFound)
}

func (c *Application) getOrganizationOverview(w http.ResponseWriter, r *http.Request) {
	org, err := c.OrganizationService().Get(r.Context(), mux.Vars(r)["organization_name"])
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	c.render("organization_get.tmpl", w, r, org)
}

func (c *Application) editOrganization(w http.ResponseWriter, r *http.Request) {
	organization, err := c.OrganizationService().Get(r.Context(), mux.Vars(r)["organization_name"])
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	c.render("organization_edit.tmpl", w, r, organization)
}

func (c *Application) updateOrganization(w http.ResponseWriter, r *http.Request) {
	var opts otf.OrganizationUpdateOptions
	if err := decode.Form(&opts, r); err != nil {
		writeError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	organization, err := c.OrganizationService().Update(r.Context(), mux.Vars(r)["organization_name"], &opts)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	flashSuccess(w, "updated organization")
	// Explicitly specify route variable for organization name because the user
	// might have updated it.
	http.Redirect(w, r, c.route("editOrganization", "organization_name", organization.Name()), http.StatusFound)
}

func (c *Application) deleteOrganization(w http.ResponseWriter, r *http.Request) {
	organizationName := mux.Vars(r)["organization_name"]
	err := c.OrganizationService().Delete(r.Context(), organizationName)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	flashSuccess(w, "deleted organization: "+organizationName)
	http.Redirect(w, r, c.route("listOrganization"), http.StatusFound)
}
