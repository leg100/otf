package html

import (
	"net/http"

	"github.com/leg100/otf"
	"github.com/leg100/otf/http/decode"
)

type newModuleStep string

const (
	newModuleConnectStep newModuleStep = "connect-vcs"
	newModuleRepoStep    newModuleStep = "select-repo"
	newModuleConfirmStep newModuleStep = "confirm-selection"
)

func (app *Application) newModule(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Step newModuleStep `schema:"step"`
	}
	var params parameters
	if err := decode.All(&params, r); err != nil {
		writeError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	switch params.Step {
	case newModuleConnectStep, "":
		app.newModuleConnect(w, r)
	case newModuleRepoStep:
		app.newModuleRepo(w, r)
	case newModuleConfirmStep:
		app.newModuleConfirm(w, r)
	}
}

func (app *Application) newModuleConnect(w http.ResponseWriter, r *http.Request) {
	org, err := decode.Param("organization_name", r)
	if err != nil {
		writeError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	providers, err := app.ListVCSProviders(r.Context(), org)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	app.render("module_new.tmpl", w, r, struct {
		Items        []*otf.VCSProvider
		Organization string
		Step         newModuleStep
	}{
		Items:        providers,
		Organization: org,
		Step:         newModuleConnectStep,
	})
}

func (app *Application) newModuleRepo(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Organization    string `schema:"organization_name,required"`
		VCSProviderID   string `schema:"vcs_provider_id,required"`
		otf.ListOptions        // paginate repos
		// TODO: filters, public/private, etc
	}
	var params parameters
	if err := decode.All(&params, r); err != nil {
		writeError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	repos, err := app.ListRepositories(r.Context(), params.VCSProviderID, params.ListOptions)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	app.render("module_new.tmpl", w, r, struct {
		*otf.RepoList
		parameters
		Step newModuleStep
	}{
		RepoList:   repos,
		parameters: params,
		Step:       newModuleRepoStep,
	})
}

func (app *Application) newModuleConfirm(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Organization  string `schema:"organization_name,required"`
		VCSProviderID string `schema:"vcs_provider_id,required"`
		Identifier    string `schema:"identifier,required"`
	}
	var params parameters
	if err := decode.All(&params, r); err != nil {
		writeError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	provider, err := app.GetVCSProvider(r.Context(), params.VCSProviderID)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	app.render("module_new.tmpl", w, r, struct {
		parameters
		Step newModuleStep
		*otf.VCSProvider
	}{
		parameters:  params,
		Step:        newModuleConfirmStep,
		VCSProvider: provider,
	})
}

func (app *Application) listModules(w http.ResponseWriter, r *http.Request) {
	var opts otf.ListModulesOptions
	if err := decode.Route(&opts, r); err != nil {
		writeError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	modules, err := app.ListModules(r.Context(), opts)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	app.render("module_list.tmpl", w, r, struct {
		Items        []*otf.Module
		Organization string
	}{
		Items:        modules,
		Organization: opts.Organization,
	})
}

func (app *Application) createModule(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Organization  string `schema:"organization_name,required"`
		VCSProviderID string `schema:"vcs_provider_id,required"`
		Identifier    string `schema:"identifier,required"`
	}
	var params parameters
	if err := decode.All(&params, r); err != nil {
		writeError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
}
