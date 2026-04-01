package ui

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/ui/helpers"
	"github.com/leg100/otf/internal/ui/paths"
)

type Handlers struct {
	Organizations                OrganizationService
	RestrictOrganizationCreation bool
}

type OrganizationService interface {
	CreateOrganization(ctx context.Context, opts organization.CreateOptions) (*organization.Organization, error)
	UpdateOrganization(ctx context.Context, name organization.Name, opts organization.UpdateOptions) (*organization.Organization, error)
	GetOrganization(ctx context.Context, name organization.Name) (*organization.Organization, error)
	ListOrganizations(ctx context.Context, opts organization.ListOptions) (*resource.Page[*organization.Organization], error)
	DeleteOrganization(ctx context.Context, name organization.Name) error
	CreateOrganizationToken(ctx context.Context, opts organization.CreateOrganizationTokenOptions) (*organization.OrganizationToken, []byte, error)
	ListOrganizationTokens(ctx context.Context, org organization.Name) ([]*organization.OrganizationToken, error)
	DeleteOrganizationToken(ctx context.Context, org organization.Name) error
}

func NewHandlers(organizations OrganizationService, restrictOrganizationCreation bool) *Handlers {
	return &Handlers{
		Organizations:                organizations,
		RestrictOrganizationCreation: restrictOrganizationCreation,
	}
}

func (h *Handlers) AddHandlers(r *mux.Router) {
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
	helpers.RenderPage(
		organizationNew(),
		"new organization",
		w,
		r,
		helpers.WithBreadcrumbs(
			helpers.Breadcrumb{Name: "organizations", Link: paths.Organizations()},
			helpers.Breadcrumb{Name: "new"},
		),
	)
}

func (h *Handlers) createOrganization(w http.ResponseWriter, r *http.Request) {
	var opts organization.CreateOptions
	if err := decode.Form(&opts, r); err != nil {
		helpers.Error(r, w, err.Error(), helpers.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	org, err := h.Organizations.CreateOrganization(r.Context(), opts)
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}

	helpers.FlashSuccess(w, "created organization: "+org.Name.String())
	http.Redirect(w, r, paths.Organization(org.Name), http.StatusFound)
}

func (h *Handlers) listOrganizations(w http.ResponseWriter, r *http.Request) {
	var opts organization.ListOptions
	if err := decode.All(&opts, r); err != nil {
		helpers.Error(r, w, err.Error(), helpers.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	organizations, err := h.Organizations.ListOrganizations(r.Context(), opts)
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}

	// Only enable the 'new organization' button if:
	// (a) RestrictOrganizationCreation is false, or
	// (b) The user has site permissions.
	subject, err := authz.SubjectFromContext(r.Context())
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}
	canCreate := !h.RestrictOrganizationCreation || subject.CanAccess(authz.CreateOrganizationAction, authz.Request{ID: resource.SiteID})

	helpers.RenderPage(
		organizationList(organizationListProps{
			Page:      organizations,
			CanCreate: canCreate,
		}),
		"organizations",
		w,
		r,
		helpers.WithBreadcrumbs(
			helpers.Breadcrumb{Name: "Organizations"},
		),
		helpers.WithContentActions(organizationListActions(canCreate)),
	)
}

func (h *Handlers) getOrganization(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Name organization.Name `schema:"name"`
	}
	if err := decode.All(&params, r); err != nil {
		helpers.Error(r, w, err.Error(), helpers.WithStatus(http.StatusUnprocessableEntity))
		return
	}
	http.Redirect(w, r, paths.Workspaces(params.Name), http.StatusFound)
}

func (h *Handlers) editOrganization(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Name organization.Name `schema:"name"`
	}
	if err := decode.All(&params, r); err != nil {
		helpers.Error(r, w, err.Error(), helpers.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	org, err := h.Organizations.GetOrganization(r.Context(), params.Name)
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}

	helpers.RenderPage(
		organizationEdit(org),
		org.Name.String(),
		w,
		r,
		helpers.WithBreadcrumbs(
			helpers.Breadcrumb{Name: "Settings"},
		),
		helpers.WithOrganization(org.Name),
	)
}

func (h *Handlers) updateOrganization(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Name            organization.Name `schema:"name,required"`
		UpdatedName     string            `schema:"new_name,required"`
		SentinelVersion string            `schema:"sentinel_version"`
	}
	if err := decode.All(&params, r); err != nil {
		helpers.Error(r, w, err.Error(), helpers.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	org, err := h.Organizations.UpdateOrganization(r.Context(), params.Name, organization.UpdateOptions{
		Name:            &params.UpdatedName,
		SentinelVersion: &params.SentinelVersion,
	})
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}

	helpers.FlashSuccess(w, "updated organization")
	http.Redirect(w, r, paths.EditOrganization(org.Name), http.StatusFound)
}

func (h *Handlers) deleteOrganization(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Name organization.Name `schema:"name"`
	}
	if err := decode.All(&params, r); err != nil {
		helpers.Error(r, w, err.Error(), helpers.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	if err := h.Organizations.DeleteOrganization(r.Context(), params.Name); err != nil {
		helpers.Error(r, w, err.Error())
		return
	}

	helpers.FlashSuccess(w, "deleted organization: "+params.Name.String())
	http.Redirect(w, r, paths.Organizations(), http.StatusFound)
}

//
// Organization tokens
//

func (h *Handlers) createOrganizationToken(w http.ResponseWriter, r *http.Request) {
	var opts organization.CreateOrganizationTokenOptions
	if err := decode.Route(&opts, r); err != nil {
		helpers.Error(r, w, err.Error(), helpers.WithStatus(http.StatusUnprocessableEntity))
		return
	}
	_, token, err := h.Organizations.CreateOrganizationToken(r.Context(), opts)
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}

	if err := helpers.TokenFlashMessage(w, token); err != nil {
		helpers.Error(r, w, err.Error())
		return
	}
	http.Redirect(w, r, paths.OrganizationToken(opts.Organization), http.StatusFound)
}

func (h *Handlers) organizationToken(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Name organization.Name `schema:"organization_name"`
	}
	if err := decode.All(&params, r); err != nil {
		helpers.Error(r, w, err.Error(), helpers.WithStatus(http.StatusUnprocessableEntity))
		return
	}
	// ListOrganizationTokens should only ever return either 0 or 1 token
	tokens, err := h.Organizations.ListOrganizationTokens(r.Context(), params.Name)
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}
	var token *organization.OrganizationToken
	if len(tokens) > 0 {
		token = tokens[0]
	}
	helpers.RenderPage(
		getToken(params.Name, token),
		params.Name.String(),
		w,
		r,
		helpers.WithBreadcrumbs(
			helpers.Breadcrumb{Name: "Organization Token"},
		),
		helpers.WithOrganization(params.Name),
	)
}

func (h *Handlers) deleteOrganizationToken(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Name organization.Name `schema:"organization_name"`
	}
	if err := decode.All(&params, r); err != nil {
		helpers.Error(r, w, err.Error(), helpers.WithStatus(http.StatusUnprocessableEntity))
		return
	}
	if err := h.Organizations.DeleteOrganizationToken(r.Context(), params.Name); err != nil {
		helpers.Error(r, w, err.Error())
		return
	}
	helpers.FlashSuccess(w, "Deleted organization token")
	http.Redirect(w, r, paths.OrganizationToken(params.Name), http.StatusFound)
}
