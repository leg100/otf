package html

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/leg100/otf/http/decode"
)

func (app *Application) newVCSProviderGithub(w http.ResponseWriter, r *http.Request) {
	app.render("vcs_provider_github_new.tmpl", w, r, organizationRequest{r})
}

func (app *Application) createVCSProviderGithub(w http.ResponseWriter, r *http.Request) {
	var opts otf.VCSProviderCreateOptions
	if err := decode.All(&opts, r); err != nil {
		writeError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	provider, err := app.CreateVCSProvider(r.Context(), opts)
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
		Items []*otf.VCSProvider
		organizationRoute
	}{
		Items:             providers,
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
