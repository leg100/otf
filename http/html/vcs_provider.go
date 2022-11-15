package html

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/leg100/otf/http/decode"
)

func (app *Application) newVCSProvider(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["cloud_name"]

	cloud, err := app.cloudDB.lookup(otf.CloudName(name))
	if err != nil {
		writeError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	tmpl := fmt.Sprintf("vcs_provider_%s_new.tmpl", name)
	app.render(tmpl, w, r, struct {
		organizationRoute
		*otf.CloudConfig
	}{
		organizationRoute: organizationRequest{r},
		CloudConfig:       cloud,
	})
}

func (app *Application) createVCSProvider(w http.ResponseWriter, r *http.Request) {
	type options struct {
		OrganizationName string        `schema:"organization_name,required"`
		Token            string        `schema:"token,required"`
		Name             string        `schema:"name,required"`
		Cloud            otf.CloudName `schema:"cloud_name,required"`
	}
	var opts options
	if err := decode.All(&opts, r); err != nil {
		writeError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	cloud, err := app.cloudDB.lookup(opts.Cloud)
	if err != nil {
		writeError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	provider, err := app.CreateVCSProvider(r.Context(), otf.VCSProviderCreateOptions{
		OrganizationName:    opts.OrganizationName,
		Token:               opts.Token,
		Name:                opts.Name,
		CloudName:           cloud.Name,
		Hostname:            cloud.Hostname,
		SkipTLSVerification: cloud.SkipTLSVerification,
		Cloud:               cloud.Cloud,
	})
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	flashSuccess(w, "created provider: "+provider.Name())
	http.Redirect(w, r, listVCSProviderPath(provider), http.StatusFound)
}

func (app *Application) listVCSProviders(w http.ResponseWriter, r *http.Request) {
	providers, err := app.ListVCSProviders(r.Context(), mux.Vars(r)["organization_name"])
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	app.render("vcs_provider_list.tmpl", w, r, struct {
		Items  []*otf.VCSProvider
		Clouds cloudDB
		organizationRoute
	}{
		Items:             providers,
		Clouds:            app.cloudDB,
		organizationRoute: organizationRequest{r},
	})
}

func (app *Application) deleteVCSProvider(w http.ResponseWriter, r *http.Request) {
	type deleteOptions struct {
		ID               string `schema:"id,required"`
		OrganizationName string `schema:"organization_name,required"`
	}
	var opts deleteOptions
	if err := decode.All(&opts, r); err != nil {
		writeError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	err := app.DeleteVCSProvider(r.Context(), opts.ID, opts.OrganizationName)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	flashSuccess(w, "deleted provider: "+opts.ID)
	http.Redirect(w, r, listVCSProviderPath(organizationRequest{r}), http.StatusFound)
}
