package ui

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/http/html"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/ui/helpers"
	"github.com/leg100/otf/internal/ui/paths"
)

// addOrganizationHandlers registers organization UI handlers with the router
func addOrganizationHandlers(r *mux.Router, h *Handlers) {
	r.HandleFunc("/organizations", h.listOrganizations).Methods("GET")
	r.HandleFunc("/organizations/new", h.newOrganization).Methods("GET")
	r.HandleFunc("/organizations/create", h.createOrganization).Methods("POST")
	r.HandleFunc("/organizations/{name}", h.getOrganization).Methods("GET")
	r.HandleFunc("/organizations/{name}/edit", h.editOrganization).Methods("GET")
	r.HandleFunc("/organizations/{name}/update", h.updateOrganization).Methods("POST")
	r.HandleFunc("/organizations/{name}/delete", h.deleteOrganization).Methods("POST")

	// organization tokens
	r.HandleFunc("/organizations/{organization_name}/tokens/show", h.organizationToken).Methods("GET")
	r.HandleFunc("/organizations/{organization_name}/tokens/delete", h.deleteOrganizationToken).Methods("POST")
	r.HandleFunc("/organizations/{organization_name}/tokens/create", h.createOrganizationToken).Methods("POST")
}

func (h *Handlers) newOrganization(w http.ResponseWriter, r *http.Request) {
	h.renderPage(
		h.templates.organizationNew(),
		"new organization",
		w,
		r,
		withBreadcrumbs(
			helpers.Breadcrumb{Name: "organizations", Link: paths.Organizations()},
			helpers.Breadcrumb{Name: "new"},
		),
	)
}

func (h *Handlers) createOrganization(w http.ResponseWriter, r *http.Request) {
	var opts organization.CreateOptions
	if err := decode.Form(&opts, r); err != nil {
		html.Error(r, w, err.Error(), html.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	org, err := h.Organizations.Create(r.Context(), opts)
	if err != nil {
		html.Error(r, w, err.Error())
		return
	}

	html.FlashSuccess(w, "created organization: "+org.Name.String())
	http.Redirect(w, r, paths.Organization(org.Name), http.StatusFound)
}

func (h *Handlers) listOrganizations(w http.ResponseWriter, r *http.Request) {
	var opts organization.ListOptions
	if err := decode.All(&opts, r); err != nil {
		html.Error(r, w, err.Error(), html.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	organizations, err := h.Organizations.List(r.Context(), opts)
	if err != nil {
		html.Error(r, w, err.Error())
		return
	}

	// Only enable the 'new organization' button if:
	// (a) RestrictOrganizationCreation is false, or
	// (b) The user has site permissions.
	subject, err := authz.SubjectFromContext(r.Context())
	if err != nil {
		html.Error(r, w, err.Error())
		return
	}
	canCreate := !h.RestrictOrganizationCreation || subject.CanAccess(authz.CreateOrganizationAction, authz.Request{ID: resource.SiteID})

	h.renderPage(
		h.templates.organizationList(organizationListProps{
			Page:      organizations,
			CanCreate: canCreate,
		}),
		"organizations",
		w,
		r,
		withBreadcrumbs(
			helpers.Breadcrumb{Name: "Organizations"},
		),
		withContentActions(organizationListActions(canCreate)),
	)
}

func (h *Handlers) getOrganization(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Name organization.Name `schema:"name"`
	}
	if err := decode.All(&params, r); err != nil {
		html.Error(r, w, err.Error(), html.WithStatus(http.StatusUnprocessableEntity))
		return
	}
	http.Redirect(w, r, paths.Workspaces(params.Name), http.StatusFound)
}

func (h *Handlers) editOrganization(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Name organization.Name `schema:"name"`
	}
	if err := decode.All(&params, r); err != nil {
		html.Error(r, w, err.Error(), html.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	org, err := h.Organizations.Get(r.Context(), params.Name)
	if err != nil {
		html.Error(r, w, err.Error())
		return
	}

	h.renderPage(
		h.templates.organizationEdit(org),
		org.Name.String(),
		w,
		r,
		withBreadcrumbs(
			helpers.Breadcrumb{Name: "Settings"},
		),
		withOrganization(org.Name),
	)
}

func (h *Handlers) updateOrganization(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Name        organization.Name `schema:"name,required"`
		UpdatedName string            `schema:"new_name,required"`
	}
	if err := decode.All(&params, r); err != nil {
		html.Error(r, w, err.Error(), html.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	org, err := h.Organizations.Update(r.Context(), params.Name, organization.UpdateOptions{
		Name: &params.UpdatedName,
	})
	if err != nil {
		html.Error(r, w, err.Error())
		return
	}

	html.FlashSuccess(w, "updated organization")
	http.Redirect(w, r, paths.EditOrganization(org.Name), http.StatusFound)
}

func (h *Handlers) deleteOrganization(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Name organization.Name `schema:"name"`
	}
	if err := decode.All(&params, r); err != nil {
		html.Error(r, w, err.Error(), html.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	if err := h.Organizations.Delete(r.Context(), params.Name); err != nil {
		html.Error(r, w, err.Error())
		return
	}

	html.FlashSuccess(w, "deleted organization: "+params.Name.String())
	http.Redirect(w, r, paths.Organizations(), http.StatusFound)
}

//
// Organization tokens
//

func (h *Handlers) createOrganizationToken(w http.ResponseWriter, r *http.Request) {
	var opts organization.CreateOrganizationTokenOptions
	if err := decode.Route(&opts, r); err != nil {
		html.Error(r, w, err.Error(), html.WithStatus(http.StatusUnprocessableEntity))
		return
	}
	_, token, err := h.Organizations.CreateToken(r.Context(), opts)
	if err != nil {
		html.Error(r, w, err.Error())
		return
	}

	if err := helpers.TokenFlashMessage(w, token); err != nil {
		html.Error(r, w, err.Error())
		return
	}
	http.Redirect(w, r, paths.OrganizationToken(opts.Organization), http.StatusFound)
}

func (h *Handlers) organizationToken(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Name organization.Name `schema:"organization_name"`
	}
	if err := decode.All(&params, r); err != nil {
		html.Error(r, w, err.Error(), html.WithStatus(http.StatusUnprocessableEntity))
		return
	}
	// ListOrganizationTokens should only ever return either 0 or 1 token
	tokens, err := h.Organizations.ListTokens(r.Context(), params.Name)
	if err != nil {
		html.Error(r, w, err.Error())
		return
	}
	var token *organization.OrganizationToken
	if len(tokens) > 0 {
		token = tokens[0]
	}
	h.renderPage(
		h.templates.getToken(params.Name, token),
		params.Name.String(),
		w,
		r,
		withBreadcrumbs(
			helpers.Breadcrumb{Name: "Organization Token"},
		),
		withOrganization(params.Name),
	)
}

func (h *Handlers) deleteOrganizationToken(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Name organization.Name `schema:"organization_name"`
	}
	if err := decode.All(&params, r); err != nil {
		html.Error(r, w, err.Error(), html.WithStatus(http.StatusUnprocessableEntity))
		return
	}
	if err := h.Organizations.DeleteToken(r.Context(), params.Name); err != nil {
		html.Error(r, w, err.Error())
		return
	}
	html.FlashSuccess(w, "Deleted organization token")
	http.Redirect(w, r, paths.OrganizationToken(params.Name), http.StatusFound)
}
