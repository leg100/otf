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

// web provides handlers for the webui
type web struct {
	otf.Signer
	otf.Renderer
	otf.VCSProviderService

	svc service
}

type newModuleStep string

const (
	newModuleConnectStep newModuleStep = "connect-vcs"
	newModuleRepoStep    newModuleStep = "select-repo"
	newModuleConfirmStep newModuleStep = "confirm-selection"
)

func (h *web) addHandlers(r *mux.Router) {
	r.HandleFunc("/organizations/{organization_name}/modules", h.listModules)
	r.HandleFunc("/organizations/{organization_name}/modules/new", h.newModule)
	r.HandleFunc("/modules/{module_id}", h.getModule)
	r.HandleFunc("/modules/{module_id}/delete", h.deleteModule)
	r.HandleFunc("/modules/create", h.createModule)
}

func (h *web) listModules(w http.ResponseWriter, r *http.Request) {
	var opts otf.ListModulesOptions
	if err := decode.All(&opts, r); err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	modules, err := h.svc.ListModules(r.Context(), opts)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.Render("module_list.tmpl", w, r, struct {
		Items        []*otf.Module
		Organization string
	}{
		Items:        modules,
		Organization: opts.Organization,
	})
}

func (h *web) getModule(w http.ResponseWriter, r *http.Request) {
	var params struct {
		ID      string `schema:"module_id,required"`
		Version string `schema:"version"`
	}
	if err := decode.All(&params, r); err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	module, err := h.svc.GetModuleByID(r.Context(), params.ID)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var tfmod *TerraformModule
	var readme template.HTML
	switch module.Status {
	case otf.ModuleStatusSetupComplete:
		tarball, err := h.svc.downloadVersion(r.Context(), otf.DownloadModuleOptions{
			ModuleVersionID: module.Version(params.Version).ID,
		})
		if err != nil {
			html.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tfmod, err = unmarshalTerraformModule(tarball)
		if err != nil {
			html.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		readme = html.MarkdownToHTML(tfmod.Readme())
	}

	h.Render("module_get.tmpl", w, r, struct {
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
		Hostname:        h.Hostname(),
	})
}

func (h *web) newModule(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Step newModuleStep `schema:"step"`
	}
	if err := decode.All(&params, r); err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	switch params.Step {
	case newModuleConnectStep, "":
		h.newModuleConnect(w, r)
	case newModuleRepoStep:
		h.newModuleRepo(w, r)
	case newModuleConfirmStep:
		h.newModuleConfirm(w, r)
	}
}

func (h *web) newModuleConnect(w http.ResponseWriter, r *http.Request) {
	org, err := decode.Param("organization_name", r)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	providers, err := h.ListVCSProviders(r.Context(), org)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.Render("module_new.tmpl", w, r, struct {
		Items        []*otf.VCSProvider
		Organization string
		Step         newModuleStep
	}{
		Items:        providers,
		Organization: org,
		Step:         newModuleConnectStep,
	})
}

func (h *web) newModuleRepo(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Organization  string `schema:"organization_name,required"`
		VCSProviderID string `schema:"vcs_provider_id,required"`
		// TODO: filters, public/private, etc
	}
	if err := decode.All(&params, r); err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	client, err := h.GetVCSClient(r.Context(), params.VCSProviderID)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	repos, err := listTerraformModuleRepos(r.Context(), client)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.Render("module_new.tmpl", w, r, struct {
		Items         []cloud.Repo
		Organization  string
		VCSProviderID string
		Step          newModuleStep
	}{
		Items:         repos,
		Organization:  params.Organization,
		VCSProviderID: params.VCSProviderID,
		Step:          newModuleRepoStep,
	})
}

func (h *web) newModuleConfirm(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Organization  string `schema:"organization_name,required"`
		VCSProviderID string `schema:"vcs_provider_id,required"`
		Identifier    string `schema:"identifier,required"`
	}
	if err := decode.All(&params, r); err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	vcsprov, err := h.GetVCSProvider(r.Context(), params.VCSProviderID)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.Render("module_new.tmpl", w, r, struct {
		Organization string
		Step         newModuleStep
		*otf.VCSProvider
	}{
		Organization: params.Organization,
		Step:         newModuleConfirmStep,
		VCSProvider:  vcsprov,
	})
}

// TODO: rename to publishModule
func (h *web) createModule(w http.ResponseWriter, r *http.Request) {
	var params struct {
		VCSProviderID string `schema:"vcs_provider_id,required"`
		Repo          string `schema:"identifier,required"`
	}
	if err := decode.All(&params, r); err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	module, err := h.svc.PublishModule(r.Context(), otf.PublishModuleOptions{
		RepoPath:      params.Repo,
		VCSProviderID: params.VCSProviderID,
	})
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	html.FlashSuccess(w, "published module: "+module.Name)
	http.Redirect(w, r, paths.Module(module.ID), http.StatusFound)
}

func (h *web) deleteModule(w http.ResponseWriter, r *http.Request) {
	id, err := decode.Param("module_id", r)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	deleted, err := h.svc.DeleteModule(r.Context(), id)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	html.FlashSuccess(w, "deleted module: "+deleted.Name)
	http.Redirect(w, r, paths.Modules(deleted.Organization), http.StatusFound)
}
