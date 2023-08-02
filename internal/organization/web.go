package organization

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/http/html"
	"github.com/leg100/otf/internal/http/html/paths"
	"github.com/leg100/otf/internal/rbac"
	"github.com/leg100/otf/internal/resource"
)

type (
	// web is the web application for organizations
	web struct {
		html.Renderer

		svc                          Service
		RestrictOrganizationCreation bool
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
}

func (a *web) new(w http.ResponseWriter, r *http.Request) {
	a.Render("organization_new.tmpl", w, html.NewSitePage(r, "new organization"))
}

func (a *web) create(w http.ResponseWriter, r *http.Request) {
	var opts OrganizationCreateOptions
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

	org, err := a.svc.UpdateOrganization(r.Context(), params.Name, OrganizationUpdateOptions{
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
