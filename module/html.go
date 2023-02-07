package module

import (
	"html/template"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/leg100/otf/cloud"
	"github.com/leg100/otf/http/decode"
	"github.com/leg100/otf/http/html"
	"github.com/leg100/otf/http/html/paths"
)

// htmlHandlers provides handlers for the webui
type htmlHandlers struct {
	otf.Signer
	otf.Renderer

	app appService
}

type newModuleStep string

const (
	newModuleConnectStep newModuleStep = "connect-vcs"
	newModuleRepoStep    newModuleStep = "select-repo"
	newModuleConfirmStep newModuleStep = "confirm-selection"
)

func (h *htmlHandlers) AddHTMLHandlers(r *mux.Router) {
	r.HandleFunc("/organizations/{organization_name}/modules", h.listModules)
	r.HandleFunc("/organizations/{organization_name}/modules/new", h.newModule)
	r.HandleFunc("/organizations/{organization_name}/modules/create", h.createModule)
	r.HandleFunc("/modules/{module_id}", h.getModule)
	r.HandleFunc("/modules/{module_id}/delete", h.deleteModule)
}

func (h *htmlHandlers) listModules(w http.ResponseWriter, r *http.Request) {
	var opts otf.ListModulesOptions
	if err := decode.All(&opts, r); err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	modules, err := h.app.ListModules(r.Context(), opts)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.Render("module_list.tmpl", w, r, struct {
		Items        []*Module
		Organization string
	}{
		Items:        modules,
		Organization: opts.Organization,
	})
}

func (app *htmlHandlers) getModule(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		ID      string `schema:"module_id,required"`
		Version string `schema:"version"`
	}
	var params parameters
	if err := decode.All(&params, r); err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	module, err := app.GetModuleByID(r.Context(), params.ID)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var tfmod *TerraformModule
	var readme template.HTML
	switch module.Status() {
	case otf.ModuleStatusSetupComplete:
		tarball, err := app.DownloadModuleVersion(r.Context(), otf.DownloadModuleOptions{
			ModuleVersionID: module.Version(params.Version).ID(),
		})
		if err != nil {
			html.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tfmod, err = UnmarshalTerraformModule(tarball)
		if err != nil {
			htmlHandlers.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		readme = markdownToHTML(tfmod.Readme())
	}

	app.render("module_get.tmpl", w, r, struct {
		*otf.Module
		TerraformModule *TerraformModule
		Readme          template.HTML
		CurrentVersion  *otf.ModuleVersion
		Hostname        string
	}{
		Module:          module,
		TerraformModule: tfmod,
		Readme:          readme,
		CurrentVersion:  module.Version(params.Version),
		Hostname:        app.Hostname(),
	})
}

func (app *htmlHandlers) newModule(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Step newModuleStep `schema:"step"`
	}
	var params parameters
	if err := decode.All(&params, r); err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
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

func (app *htmlHandlers) newModuleConnect(w http.ResponseWriter, r *http.Request) {
	org, err := decode.Param("organization_name", r)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	providers, err := app.ListVCSProviders(r.Context(), org)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	app.Render("module_new.tmpl", w, r, struct {
		Items        []*otf.VCSProvider
		Organization string
		Step         newModuleStep
	}{
		Items:        providers,
		Organization: org,
		Step:         newModuleConnectStep,
	})
}

func (app *htmlHandlers) newModuleRepo(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Organization  string `schema:"organization_name,required"`
		VCSProviderID string `schema:"vcs_provider_id,required"`
		// TODO: filters, public/private, etc
	}
	var params parameters
	if err := decode.All(&params, r); err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	repos, err := otf.ListModuleRepositories(r.Context(), app, params.VCSProviderID)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	app.Render("module_new.tmpl", w, r, struct {
		Items []cloud.Repo
		parameters
		Step newModuleStep
	}{
		Items:      repos,
		parameters: params,
		Step:       newModuleRepoStep,
	})
}

func (app *htmlHandlers) newModuleConfirm(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Organization  string `schema:"organization_name,required"`
		VCSProviderID string `schema:"vcs_provider_id,required"`
		Identifier    string `schema:"identifier,required"`
	}
	var params parameters
	if err := decode.All(&params, r); err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	provider, err := app.GetVCSProvider(r.Context(), params.VCSProviderID)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	app.Render("module_new.tmpl", w, r, struct {
		parameters
		Step newModuleStep
		otf.VCSProvider
	}{
		parameters:  params,
		Step:        newModuleConfirmStep,
		VCSProvider: provider,
	})
}

func (app *htmlHandlers) createModule(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Organization  string `schema:"organization_name,required"`
		VCSProviderID string `schema:"vcs_provider_id,required"`
		Identifier    string `schema:"identifier,required"`
	}
	var params parameters
	if err := decode.All(&params, r); err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	org, err := app.GetOrganization(r.Context(), params.Organization)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	module, err := app.PublishModule(r.Context(), otf.PublishModuleOptions{
		Identifier:   params.Identifier,
		ProviderID:   params.VCSProviderID,
		Organization: org,
	})
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	html.FlashSuccess(w, "published module: "+module.Name())
	http.Redirect(w, r, paths.Module(module.ID()), http.StatusFound)
}

func (app *htmlHandlers) deleteModule(w http.ResponseWriter, r *http.Request) {
	id, err := decode.Param("module_id", r)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	deleted, err := app.DeleteModule(r.Context(), id)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	html.FlashSuccess(w, "deleted module: "+deleted.Name())
	http.Redirect(w, r, paths.Modules(deleted.Organization()), http.StatusFound)
}
