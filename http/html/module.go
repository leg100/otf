package html

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf"
)

func (app *Application) listModules(w http.ResponseWriter, r *http.Request) {
	providers, err := app.ListVCSProviders(r.Context(), mux.Vars(r)["organization_name"])
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	app.render("module_list.tmpl", w, r, struct {
		Items        []*otf.VCSProvider
		CloudConfigs []otf.CloudConfig
		organizationRoute
	}{
		Items:             providers,
		CloudConfigs:      app.ListCloudConfigs(),
		organizationRoute: organizationRequest{r},
	})
}

func (app *Application) listModuleVCSProviders(w http.ResponseWriter, r *http.Request) {
	providers, err := app.ListVCSProviders(r.Context(), mux.Vars(r)["organization_name"])
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	app.render("workspace_vcs_provider_list.tmpl", w, r, struct {
		Items []*otf.VCSProvider
		workspaceRoute
	}{
		Items:          providers,
		workspaceRoute: workspaceRequest{r},
	})
}
