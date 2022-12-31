package html

import (
	"net/http"

	"github.com/leg100/otf"
	otfhttp "github.com/leg100/otf/http"
	"github.com/leg100/otf/http/decode"
	"github.com/leg100/otf/http/html/paths"
)

type newModuleStep string

const (
	newModuleConnectStep newModuleStep = "connect-vcs"
	newModuleRepoStep    newModuleStep = "select-repo"
	newModuleConfirmStep newModuleStep = "confirm-selection"
)

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

func (app *Application) getModule(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		ID      string `schema:"module_id,required"`
		Version string `schema:"version"`
	}
	var params parameters
	if err := decode.All(&params, r); err != nil {
		writeError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	module, err := app.GetModuleByID(r.Context(), params.ID)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	app.render("module_get.tmpl", w, r, struct {
		*otf.Module
		CurrentVersion *otf.ModuleVersion
	}{
		Module:         module,
		CurrentVersion: module.Version(params.Version),
	})
}

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
		Organization  string `schema:"organization_name,required"`
		VCSProviderID string `schema:"vcs_provider_id,required"`
		// TODO: filters, public/private, etc
	}
	var params parameters
	if err := decode.All(&params, r); err != nil {
		writeError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	repos, err := otf.ListModuleRepositories(r.Context(), app, params.VCSProviderID)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	app.render("module_new.tmpl", w, r, struct {
		Items []*otf.Repo
		parameters
		Step newModuleStep
	}{
		Items:      repos,
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

	org, err := app.GetOrganization(r.Context(), params.Organization)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	module, err := app.PublishModule(r.Context(), otf.PublishModuleOptions{
		Identifier:   params.Identifier,
		ProviderID:   params.VCSProviderID,
		OTFHost:      otfhttp.ExternalHost(r),
		Organization: org,
	})
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	flashSuccess(w, "published module: "+module.Name())
	http.Redirect(w, r, paths.Module(module.ID()), http.StatusFound)
}

func (app *Application) deleteModule(w http.ResponseWriter, r *http.Request) {
	id, err := decode.Param("module_id", r)
	if err != nil {
		writeError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	module, err := app.GetModuleByID(r.Context(), id)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = app.DeleteModule(r.Context(), id)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	flashSuccess(w, "deleted module: "+module.Name())
	http.Redirect(w, r, paths.Modules(module.Organization().Name()), http.StatusFound)
}
