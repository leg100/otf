package html

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/leg100/otf/cloud"
	"github.com/leg100/otf/http/decode"
	"github.com/leg100/otf/http/html/paths"
)

func (app *Application) newVCSProvider(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Organization string `schema:"organization_name,required"`
		Cloud        string `schema:"cloud,required"`
	}
	var params parameters
	if err := decode.All(&params, r); err != nil {
		Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	tmpl := fmt.Sprintf("vcs_provider_%s_new.tmpl", params.Cloud)
	app.Render(tmpl, w, r, params)
}

func (app *Application) createVCSProvider(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		OrganizationName string `schema:"organization_name,required"`
		Token            string `schema:"token,required"`
		Name             string `schema:"name,required"`
		Cloud            string `schema:"cloud,required"`
	}
	var params parameters
	if err := decode.All(&params, r); err != nil {
		Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	provider, err := app.CreateVCSProvider(r.Context(), otf.VCSProviderCreateOptions{
		Organization: params.OrganizationName,
		Token:        params.Token,
		Name:         params.Name,
		Cloud:        params.Cloud,
	})
	if err != nil {
		Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	FlashSuccess(w, "created provider: "+provider.Name())
	http.Redirect(w, r, paths.VCSProviders(provider.Organization()), http.StatusFound)
}

func (app *Application) listVCSProviders(w http.ResponseWriter, r *http.Request) {
	org, err := app.GetOrganization(r.Context(), mux.Vars(r)["organization_name"])
	if err != nil {
		Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	providers, err := app.ListVCSProviders(r.Context(), mux.Vars(r)["organization_name"])
	if err != nil {
		Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	app.Render("vcs_provider_list.tmpl", w, r, struct {
		Items        []otf.VCSProvider
		CloudConfigs []cloud.Config
		*otf.Organization
	}{
		Items:        providers,
		CloudConfigs: app.ListCloudConfigs(),
		Organization: org,
	})
}

func (app *Application) deleteVCSProvider(w http.ResponseWriter, r *http.Request) {
	id, err := decode.Param("vcs_provider_id", r)
	if err != nil {
		Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	provider, err := app.DeleteVCSProvider(r.Context(), id)
	if err != nil {
		Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	FlashSuccess(w, "deleted provider: "+provider.Name())
	http.Redirect(w, r, paths.VCSProviders(provider.Organization()), http.StatusFound)
}
