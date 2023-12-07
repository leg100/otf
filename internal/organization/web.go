package organization

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/http/html"
	"github.com/leg100/otf/internal/http/html/paths"
	"github.com/leg100/otf/internal/rbac"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/tokens"
)

type (
	// web is the web application for organizations
	web struct {
		html.Renderer

		svc                          webService
		RestrictOrganizationCreation bool
	}

	// webService provides the web app with access to organizations
	webService interface {
		CreateOrganization(ctx context.Context, opts CreateOptions) (*Organization, error)
		UpdateOrganization(ctx context.Context, name string, opts UpdateOptions) (*Organization, error)
		GetOrganization(ctx context.Context, name string) (*Organization, error)
		ListOrganizations(ctx context.Context, opts ListOptions) (*resource.Page[*Organization], error)
		DeleteOrganization(ctx context.Context, name string) error

		CreateOrganizationToken(ctx context.Context, opts CreateOrganizationTokenOptions) (*OrganizationToken, []byte, error)
		ListOrganizationTokens(ctx context.Context, organization string) ([]*OrganizationToken, error)
		DeleteOrganizationToken(ctx context.Context, organization string) error
	}

	// OrganizationPage contains data shared by all organization-based pages.
	OrganizationPage struct {
		html.SitePage

		Organization string
	}
)

func NewPage(r *http.Request, title, organization string) OrganizationPage {
	sitePage := html.NewSitePage(r, title)
	sitePage.CurrentOrganization = organization
	return OrganizationPage{
		Organization: organization,
		SitePage:     sitePage,
	}
}

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
	a.Render("organization_new.tmpl", w, html.NewSitePage(r, "new organization"))
}

func (a *web) create(w http.ResponseWriter, r *http.Request) {
	var opts CreateOptions
	if err := decode.Form(&opts, r); err != nil {
		a.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	org, err := a.svc.CreateOrganization(r.Context(), opts)
	if err == internal.ErrResourceAlreadyExists {
		html.FlashError(w, "organization already exists: "+*opts.Name)
		http.Redirect(w, r, paths.NewOrganization(), http.StatusFound)
		return
	}
	if err != nil {
		a.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	html.FlashSuccess(w, "created organization: "+org.Name)
	http.Redirect(w, r, paths.Organization(org.Name), http.StatusFound)
}

func (a *web) list(w http.ResponseWriter, r *http.Request) {
	var params struct {
		PageNumber int `schema:"page[number]"`
	}
	if err := decode.All(&params, r); err != nil {
		a.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	organizations, err := a.svc.ListOrganizations(r.Context(), ListOptions{
		PageOptions: resource.PageOptions{
			PageNumber: params.PageNumber,
			PageSize:   html.PageSize,
		},
	})
	if err != nil {
		a.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Only enable the 'new organization' button if:
	// (a) RestrictOrganizationCreation is false, or
	// (b) The user has site permissions.
	subject, err := internal.SubjectFromContext(r.Context())
	if err != nil {
		a.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var canCreate bool
	if !a.RestrictOrganizationCreation || subject.CanAccessSite(rbac.CreateOrganizationAction) {
		canCreate = true
	}

	a.Render("organization_list.tmpl", w, struct {
		html.SitePage
		*resource.Page[*Organization]
		CanCreate bool
	}{
		SitePage:  html.NewSitePage(r, "organizations"),
		Page:      organizations,
		CanCreate: canCreate,
	})
}

func (a *web) get(w http.ResponseWriter, r *http.Request) {
	name, err := decode.Param("name", r)
	if err != nil {
		a.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	org, err := a.svc.GetOrganization(r.Context(), name)
	if err != nil {
		a.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	a.Render("organization_get.tmpl", w, struct {
		OrganizationPage
		*Organization
	}{
		OrganizationPage: NewPage(r, org.Name, org.Name),
		Organization:     org,
	})
}

func (a *web) edit(w http.ResponseWriter, r *http.Request) {
	name, err := decode.Param("name", r)
	if err != nil {
		a.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	org, err := a.svc.GetOrganization(r.Context(), name)
	if err != nil {
		a.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	a.Render("organization_edit.tmpl", w, struct {
		OrganizationPage
		*Organization
	}{
		OrganizationPage: NewPage(r, org.Name, org.Name),
		Organization:     org,
	})
}

func (a *web) update(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Name        string `schema:"name,required"`
		UpdatedName string `schema:"new_name,required"`
	}
	if err := decode.All(&params, r); err != nil {
		a.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	org, err := a.svc.UpdateOrganization(r.Context(), params.Name, UpdateOptions{
		Name: &params.UpdatedName,
	})
	if err != nil {
		a.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	html.FlashSuccess(w, "updated organization")
	http.Redirect(w, r, paths.EditOrganization(org.Name), http.StatusFound)
}

func (a *web) delete(w http.ResponseWriter, r *http.Request) {
	organization, err := decode.Param("name", r)
	if err != nil {
		a.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	err = a.svc.DeleteOrganization(r.Context(), organization)
	if err != nil {
		a.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	html.FlashSuccess(w, "deleted organization: "+organization)
	http.Redirect(w, r, paths.Organizations(), http.StatusFound)
}

//
// Organization tokens
//

func (a *web) createOrganizationToken(w http.ResponseWriter, r *http.Request) {
	var opts CreateOrganizationTokenOptions
	if err := decode.Route(&opts, r); err != nil {
		a.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	_, token, err := a.svc.CreateOrganizationToken(r.Context(), opts)
	if err != nil {
		a.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tokens.TokenFlashMessage(a, w, token); err != nil {
		a.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, paths.OrganizationToken(opts.Organization), http.StatusFound)
}

func (a *web) organizationToken(w http.ResponseWriter, r *http.Request) {
	org, err := decode.Param("organization_name", r)
	if err != nil {
		a.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	// ListOrganizationTokens should only ever return either 0 or 1 token
	tokens, err := a.svc.ListOrganizationTokens(r.Context(), org)
	if err != nil {
		a.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var token *OrganizationToken
	if len(tokens) > 0 {
		token = tokens[0]
	}
	a.Render("organization_token.tmpl", w, struct {
		OrganizationPage
		Token *OrganizationToken
	}{
		OrganizationPage: NewPage(r, org, org),
		Token:            token,
	})
}

func (a *web) deleteOrganizationToken(w http.ResponseWriter, r *http.Request) {
	organization, err := decode.Param("organization_name", r)
	if err != nil {
		a.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	if err := a.svc.DeleteOrganizationToken(r.Context(), organization); err != nil {
		a.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	html.FlashSuccess(w, "Deleted organization token")
	http.Redirect(w, r, paths.OrganizationToken(organization), http.StatusFound)
}
