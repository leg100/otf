package ui

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/http/html"
	"github.com/leg100/otf/internal/http/html/paths"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/ui/helpers"
)

type (
	// organizationHandlers is the web application for organizations
	organizationHandlers struct {
		svc              organizationClient
		RestrictCreation bool
	}

	// organizationClient provides the web app with access to organizations
	organizationClient interface {
		Create(ctx context.Context, opts organization.CreateOptions) (*organization.Organization, error)
		Update(ctx context.Context, name organization.Name, opts organization.UpdateOptions) (*organization.Organization, error)
		Get(ctx context.Context, name organization.Name) (*organization.Organization, error)
		List(ctx context.Context, opts organization.ListOptions) (*resource.Page[*organization.Organization], error)
		Delete(ctx context.Context, name organization.Name) error

		CreateToken(ctx context.Context, opts organization.CreateOrganizationTokenOptions) (*organization.OrganizationToken, []byte, error)
		ListTokens(ctx context.Context, org organization.Name) ([]*organization.OrganizationToken, error)
		DeleteToken(ctx context.Context, org organization.Name) error
	}
)

// addOrganizationHandlers registers organization UI handlers with the router
func addOrganizationHandlers(r *mux.Router, svc organizationClient, restrictCreation bool) {
	h := &organizationHandlers{
		svc:              svc,
		RestrictCreation: restrictCreation,
	}

	r.HandleFunc("/organizations", h.list).Methods("GET")
	r.HandleFunc("/organizations/new", h.new).Methods("GET")
	r.HandleFunc("/organizations/create", h.create).Methods("POST")
	r.HandleFunc("/organizations/{name}", h.get).Methods("GET")
	r.HandleFunc("/organizations/{name}/edit", h.edit).Methods("GET")
	r.HandleFunc("/organizations/{name}/update", h.update).Methods("POST")
	r.HandleFunc("/organizations/{name}/delete", h.delete).Methods("POST")

	// organization tokens
	r.HandleFunc("/organizations/{organization_name}/tokens/show", h.organizationToken).Methods("GET")
	r.HandleFunc("/organizations/{organization_name}/tokens/delete", h.deleteOrganizationToken).Methods("POST")
	r.HandleFunc("/organizations/{organization_name}/tokens/create", h.createOrganizationToken).Methods("POST")
}

func (a *organizationHandlers) new(w http.ResponseWriter, r *http.Request) {
	html.Render(organizationNew(), w, r)
}

func (a *organizationHandlers) create(w http.ResponseWriter, r *http.Request) {
	var opts organization.CreateOptions
	if err := decode.Form(&opts, r); err != nil {
		html.Error(r, w, err.Error(), html.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	org, err := a.svc.Create(r.Context(), opts)
	if err != nil {
		html.Error(r, w, err.Error())
		return
	}

	html.FlashSuccess(w, "created organization: "+org.Name.String())
	http.Redirect(w, r, paths.Organization(org.Name), http.StatusFound)
}

func (a *organizationHandlers) list(w http.ResponseWriter, r *http.Request) {
	var opts organization.ListOptions
	if err := decode.All(&opts, r); err != nil {
		html.Error(r, w, err.Error(), html.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	organizations, err := a.svc.List(r.Context(), opts)
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
	canCreate := !a.RestrictCreation || subject.CanAccess(authz.CreateOrganizationAction, authz.Request{ID: resource.SiteID})

	props := organizationListProps{
		Page:      organizations,
		CanCreate: canCreate,
	}
	html.Render(organizationList(props), w, r)
}

func (a *organizationHandlers) get(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Name organization.Name `schema:"name"`
	}
	if err := decode.All(&params, r); err != nil {
		html.Error(r, w, err.Error(), html.WithStatus(http.StatusUnprocessableEntity))
		return
	}
	http.Redirect(w, r, paths.Workspaces(params.Name), http.StatusFound)
}

func (a *organizationHandlers) edit(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Name organization.Name `schema:"name"`
	}
	if err := decode.All(&params, r); err != nil {
		html.Error(r, w, err.Error(), html.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	org, err := a.svc.Get(r.Context(), params.Name)
	if err != nil {
		html.Error(r, w, err.Error())
		return
	}

	html.Render(organizationEdit(org), w, r)
}

func (a *organizationHandlers) update(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Name        organization.Name `schema:"name,required"`
		UpdatedName string            `schema:"new_name,required"`
	}
	if err := decode.All(&params, r); err != nil {
		html.Error(r, w, err.Error(), html.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	org, err := a.svc.Update(r.Context(), params.Name, organization.UpdateOptions{
		Name: &params.UpdatedName,
	})
	if err != nil {
		html.Error(r, w, err.Error())
		return
	}

	html.FlashSuccess(w, "updated organization")
	http.Redirect(w, r, paths.EditOrganization(org.Name), http.StatusFound)
}

func (a *organizationHandlers) delete(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Name organization.Name `schema:"name"`
	}
	if err := decode.All(&params, r); err != nil {
		html.Error(r, w, err.Error(), html.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	if err := a.svc.Delete(r.Context(), params.Name); err != nil {
		html.Error(r, w, err.Error())
		return
	}

	html.FlashSuccess(w, "deleted organization: "+params.Name.String())
	http.Redirect(w, r, paths.Organizations(), http.StatusFound)
}

//
// Organization tokens
//

func (a *organizationHandlers) createOrganizationToken(w http.ResponseWriter, r *http.Request) {
	var opts organization.CreateOrganizationTokenOptions
	if err := decode.Route(&opts, r); err != nil {
		html.Error(r, w, err.Error(), html.WithStatus(http.StatusUnprocessableEntity))
		return
	}
	_, token, err := a.svc.CreateToken(r.Context(), opts)
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

func (a *organizationHandlers) organizationToken(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Name organization.Name `schema:"organization_name"`
	}
	if err := decode.All(&params, r); err != nil {
		html.Error(r, w, err.Error(), html.WithStatus(http.StatusUnprocessableEntity))
		return
	}
	// ListOrganizationTokens should only ever return either 0 or 1 token
	tokens, err := a.svc.ListTokens(r.Context(), params.Name)
	if err != nil {
		html.Error(r, w, err.Error())
		return
	}
	var token *organization.OrganizationToken
	if len(tokens) > 0 {
		token = tokens[0]
	}
	html.Render(getToken(params.Name, token), w, r)
}

func (a *organizationHandlers) deleteOrganizationToken(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Name organization.Name `schema:"organization_name"`
	}
	if err := decode.All(&params, r); err != nil {
		html.Error(r, w, err.Error(), html.WithStatus(http.StatusUnprocessableEntity))
		return
	}
	if err := a.svc.DeleteToken(r.Context(), params.Name); err != nil {
		html.Error(r, w, err.Error())
		return
	}
	html.FlashSuccess(w, "Deleted organization token")
	http.Redirect(w, r, paths.OrganizationToken(params.Name), http.StatusFound)
}
