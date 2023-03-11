package vcsprovider

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/leg100/otf/cloud"
	"github.com/leg100/otf/http/decode"
	"github.com/leg100/otf/http/html"
	"github.com/leg100/otf/http/html/paths"
)

type web struct {
	otf.Renderer
	workspace.Service
	cloud.Service

	app service
}

func (a *web) AddHandlers(r *mux.Router) {
	r.HandleFunc("/organizations/{organization_name}/vcs-providers", a.list)
	r.HandleFunc("/organizations/{organization_name}/vcs-providers/new", a.new)
	r.HandleFunc("/organizations/{organization_name}/vcs-providers/create", a.create)
	r.HandleFunc("/vcs-providers/{vcs_provider_id}/delete", a.delete)
}

func (a *web) new(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Organization string `schema:"organization_name,required"`
		Cloud        string `schema:"cloud,required"`
	}
	var params parameters
	if err := decode.All(&params, r); err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	tmpl := fmt.Sprintf("vcs_provider_%s_new.tmpl", params.Cloud)
	a.Render(tmpl, w, r, params)
}

func (a *web) create(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		OrganizationName string `schema:"organization_name,required"`
		Token            string `schema:"token,required"`
		Name             string `schema:"name,required"`
		Cloud            string `schema:"cloud,required"`
	}
	var params parameters
	if err := decode.All(&params, r); err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	provider, err := a.app.create(r.Context(), createOptions{
		Organization: params.OrganizationName,
		Token:        params.Token,
		Name:         params.Name,
		Cloud:        params.Cloud,
	})
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	html.FlashSuccess(w, "created provider: "+provider.Name)
	http.Redirect(w, r, paths.VCSProviders(provider.Organization), http.StatusFound)
}

func (a *web) list(w http.ResponseWriter, r *http.Request) {
	organization, err := decode.Param("organization_name", r)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	providers, err := a.app.list(r.Context(), organization)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	a.Render("vcs_provider_list.tmpl", w, r, struct {
		Items        []*otf.VCSProvider
		CloudConfigs []cloud.Config
		Organization string
	}{
		Items:        providers,
		CloudConfigs: a.ListCloudConfigs(),
		Organization: organization,
	})
}

func (a *web) delete(w http.ResponseWriter, r *http.Request) {
	id, err := decode.Param("vcs_provider_id", r)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	provider, err := a.app.delete(r.Context(), id)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	html.FlashSuccess(w, "deleted provider: "+provider.Name)
	http.Redirect(w, r, paths.VCSProviders(provider.Organization), http.StatusFound)
}
