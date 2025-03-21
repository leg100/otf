package organization

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/http/html"
	"github.com/leg100/otf/internal/http/html/components"
	"github.com/leg100/otf/internal/http/html/paths"
	"github.com/leg100/otf/internal/resource"
)

type (
	// web is the web application for organizations
	web struct {
		svc              webService
		RestrictCreation bool
	}

	// webService provides the web app with access to organizations
	webService interface {
		Create(ctx context.Context, opts CreateOptions) (*Organization, error)
		Update(ctx context.Context, name resource.OrganizationName, opts UpdateOptions) (*Organization, error)
		Get(ctx context.Context, name resource.OrganizationName) (*Organization, error)
		List(ctx context.Context, opts ListOptions) (*resource.Page[*Organization], error)
		Delete(ctx context.Context, name resource.OrganizationName) error

		CreateToken(ctx context.Context, opts CreateOrganizationTokenOptions) (*OrganizationToken, []byte, error)
		ListTokens(ctx context.Context, organization resource.OrganizationName) ([]*OrganizationToken, error)
		DeleteToken(ctx context.Context, organization resource.OrganizationName) error
	}
)

func (a *web) addHandlers(r *mux.Router) {
	r = html.UIRouter(r)

	r.HandleFunc("/organizations", a.list).Methods("GET")
	r.HandleFunc("/organizations/new", a.new).Methods("GET")
	r.HandleFunc("/organizations/create", a.create).Methods("POST")
	r.HandleFunc("/organizations/{name}", a.get).Methods("GET")
	r.HandleFunc("/organizations/{name}/edit", a.edit).Methods("GET")
	r.HandleFunc("/organizations/{name}/update", a.update).Methods("POST")
	r.HandleFunc("/organizations/{name}/delete", a.delete).Methods("POST")

	// organization tokens
	r.HandleFunc("/organizations/{organization_name}/tokens/show", a.organizationToken).Methods("GET")
	r.HandleFunc("/organizations/{organization_name}/tokens/delete", a.deleteOrganizationToken).Methods("POST")
	r.HandleFunc("/organizations/{organization_name}/tokens/create", a.createOrganizationToken).Methods("POST")
}

func (a *web) new(w http.ResponseWriter, r *http.Request) {
	html.Render(new(), w, r)
}

func (a *web) create(w http.ResponseWriter, r *http.Request) {
	var opts CreateOptions
	if err := decode.Form(&opts, r); err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	org, err := a.svc.Create(r.Context(), opts)
	if err == internal.ErrResourceAlreadyExists {
		html.FlashError(w, "organization already exists: "+*opts.Name)
		http.Redirect(w, r, paths.NewOrganization(), http.StatusFound)
		return
	}
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	html.FlashSuccess(w, "created organization: "+org.Name.String())
	http.Redirect(w, r, paths.Organization(org.Name.String()), http.StatusFound)
}

func (a *web) list(w http.ResponseWriter, r *http.Request) {
	var opts ListOptions
	if err := decode.All(&opts, r); err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	organizations, err := a.svc.List(r.Context(), opts)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Only enable the 'new organization' button if:
	// (a) RestrictOrganizationCreation is false, or
	// (b) The user has site permissions.
	subject, err := authz.SubjectFromContext(r.Context())
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	canCreate := !a.RestrictCreation || subject.CanAccess(authz.CreateOrganizationAction, authz.AccessRequest{ID: resource.SiteID})

	props := listProps{
		Page:      organizations,
		CanCreate: canCreate,
	}
	html.Render(list(props), w, r)
}

func (a *web) get(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Name resource.OrganizationName `schema:"name"`
	}
	if err := decode.All(&params, r); err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	org, err := a.svc.Get(r.Context(), params.Name)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	html.Render(get(org), w, r)
}

func (a *web) edit(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Name resource.OrganizationName `schema:"name"`
	}
	if err := decode.All(&params, r); err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	org, err := a.svc.Get(r.Context(), params.Name)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	html.Render(edit(org), w, r)
}

func (a *web) update(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Name        resource.OrganizationName `schema:"name,required"`
		UpdatedName string                    `schema:"new_name,required"`
	}
	if err := decode.All(&params, r); err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	org, err := a.svc.Update(r.Context(), params.Name, UpdateOptions{
		Name: &params.UpdatedName,
	})
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	html.FlashSuccess(w, "updated organization")
	http.Redirect(w, r, paths.EditOrganization(org.Name.String()), http.StatusFound)
}

func (a *web) delete(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Name resource.OrganizationName `schema:"name"`
	}
	if err := decode.All(&params, r); err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	if err := a.svc.Delete(r.Context(), params.Name); err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	html.FlashSuccess(w, "deleted organization: "+params.Name.String())
	http.Redirect(w, r, paths.Organizations(), http.StatusFound)
}

//
// Organization tokens
//

func (a *web) createOrganizationToken(w http.ResponseWriter, r *http.Request) {
	var opts CreateOrganizationTokenOptions
	if err := decode.Route(&opts, r); err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	_, token, err := a.svc.CreateToken(r.Context(), opts)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := components.TokenFlashMessage(w, token); err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, paths.OrganizationToken(opts.Organization.String()), http.StatusFound)
}

func (a *web) organizationToken(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Name resource.OrganizationName `schema:"organization_name"`
	}
	if err := decode.All(&params, r); err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	// ListOrganizationTokens should only ever return either 0 or 1 token
	tokens, err := a.svc.ListTokens(r.Context(), params.Name)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var token *OrganizationToken
	if len(tokens) > 0 {
		token = tokens[0]
	}
	html.Render(getToken(params.Name, token), w, r)
}

func (a *web) deleteOrganizationToken(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Name resource.OrganizationName `schema:"organization_name"`
	}
	if err := decode.All(&params, r); err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	if err := a.svc.DeleteToken(r.Context(), params.Name); err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	html.FlashSuccess(w, "Deleted organization token")
	http.Redirect(w, r, paths.OrganizationToken(params.Name.String()), http.StatusFound)
}
