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

	svc Service
}

func (a *web) addHandlers(r *mux.Router) {
	r = html.UIRouter(r)

	r.HandleFunc("/organizations", a.list).Methods("GET")
	r.HandleFunc("/organizations/{name}", a.get).Methods("GET")
	r.HandleFunc("/organizations/{name}/edit", a.edit).Methods("GET")
	r.HandleFunc("/organizations/{name}/update", a.update).Methods("POST")
	r.HandleFunc("/organizations/{name}/delete", a.delete).Methods("POST")
}

func (a *web) list(w http.ResponseWriter, r *http.Request) {
	var opts OrganizationListOptions
	if err := decode.Query(&opts, r.URL.Query()); err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	organizations, err := a.svc.ListOrganizations(r.Context(), opts)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	a.Render("organization_list.tmpl", w, r, struct {
		*OrganizationList
		OrganizationListOptions
	}{
		OrganizationList:        organizations,
		OrganizationListOptions: opts,
	})
}

func (a *web) get(w http.ResponseWriter, r *http.Request) {
	name, err := decode.Param("name", r)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	org, err := a.svc.GetOrganization(r.Context(), name)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	a.Render("organization_get.tmpl", w, r, org)
}

func (a *web) edit(w http.ResponseWriter, r *http.Request) {
	name, err := decode.Param("name", r)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	organization, err := a.svc.GetOrganization(r.Context(), name)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	a.Render("organization_edit.tmpl", w, r, organization)
}

func (a *web) update(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Options OrganizationUpdateOptions
		Name    string `schema:"name,required"`
	}
	if err := decode.All(&params, r); err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	org, err := a.svc.UpdateOrganization(r.Context(), params.Name, params.Options)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	html.FlashSuccess(w, "updated organization")
	http.Redirect(w, r, paths.EditOrganization(org.Name), http.StatusFound)
}

func (a *web) delete(w http.ResponseWriter, r *http.Request) {
	organization, err := decode.Param("name", r)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	err = a.svc.DeleteOrganization(r.Context(), organization)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	html.FlashSuccess(w, "deleted organization: "+organization)
	http.Redirect(w, r, paths.Organizations(), http.StatusFound)
}
