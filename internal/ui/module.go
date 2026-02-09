package ui

import (
	"errors"
	"html/template"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/http/html"
	"github.com/leg100/otf/internal/module"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/ui/helpers"
	"github.com/leg100/otf/internal/ui/paths"
	"github.com/leg100/otf/internal/vcs"
)

// addModuleHandlers registers module UI handlers with the router
func addModuleHandlers(r *mux.Router, h *Handlers) {
	r.HandleFunc("/organizations/{organization_name}/modules", h.listModules).Methods("GET")
	r.HandleFunc("/organizations/{organization_name}/modules/new", h.newModule).Methods("GET")
	r.HandleFunc("/organizations/{organization_name}/modules/create", h.publishModule).Methods("POST")

	r.HandleFunc("/modules/{vcs_provider_id}/connect", h.connectModule).Methods("GET")
	r.HandleFunc("/modules/{module_id}", h.getModule).Methods("GET")
	r.HandleFunc("/modules/{module_id}/delete", h.deleteModule).Methods("POST")
}

func (h *Handlers) listModules(w http.ResponseWriter, r *http.Request) {
	var params struct {
		module.ListOptions
		resource.PageOptions
		ProviderFilterVisible bool `schema:"provider_filter_visible"`
	}
	if err := decode.All(&params, r); err != nil {
		html.Error(r, w, err.Error(), html.WithStatus(http.StatusUnprocessableEntity))
		return
	}
	modules, err := h.Modules.ListModules(r.Context(), params.ListOptions)
	if err != nil {
		html.Error(r, w, err.Error())
		return
	}
	providers, err := h.Modules.ListProviders(r.Context(), params.Organization)
	if err != nil {
		html.Error(r, w, err.Error())
		return
	}

	props := moduleListProps{
		organization:          params.Organization,
		page:                  resource.NewPage(modules, params.PageOptions, nil),
		canPublishModule:      h.Authorizer.CanAccess(r.Context(), authz.CreateModuleAction, params.Organization),
		providerFilterVisible: params.ProviderFilterVisible,
		allProviders:          providers,
		selectedProviders:     params.Providers,
	}
	h.renderPage(
		h.templates.moduleList(props),
		"Modules",
		w,
		r,
		withOrganization(params.Organization),
		withBreadcrumbs(
			helpers.Breadcrumb{Name: "Modules"},
		),
		withContentActions(moduleListActions(props)),
	)
}

func (h *Handlers) getModule(w http.ResponseWriter, r *http.Request) {
	var params struct {
		ID      resource.TfeID `schema:"module_id,required"`
		Version *string        `schema:"version"`
	}
	if err := decode.All(&params, r); err != nil {
		html.Error(r, w, err.Error(), html.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	mod, err := h.Modules.GetModuleByID(r.Context(), params.ID)
	if err != nil {
		html.Error(r, w, err.Error())
		return
	}

	var (
		modver *module.ModuleVersion
		tfmod  *module.TerraformModule
		readme template.HTML
	)
	if params.Version != nil {
		modver = mod.Version(*params.Version)
	} else {
		modver = mod.Latest()
	}
	if modver != nil {
		tfmod, err = h.Modules.GetModuleInfo(r.Context(), modver.ID)
		if err != nil {
			html.Error(r, w, err.Error())
			return
		}
	}

	switch mod.Status {
	case module.ModuleStatusSetupComplete:
		if tfmod != nil {
			readme = html.MarkdownToHTML(tfmod.GetReadme())
		}
	}

	h.renderPage(
		h.templates.moduleGet(moduleGetProps{
			module:          mod,
			terraformModule: tfmod,
			readme:          readme,
			currentVersion:  modver,
			hostname:        h.HostnameService.Hostname(),
		}),
		"modules",
		w,
		r,
		withOrganization(mod.Organization),
		withBreadcrumbs(
			helpers.Breadcrumb{Name: "Modules", Link: paths.Modules(mod.Organization)},
			helpers.Breadcrumb{Name: mod.Name},
		),
	)
}

func (h *Handlers) newModule(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Organization organization.Name `schema:"organization_name"`
	}
	if err := decode.All(&params, r); err != nil {
		html.Error(r, w, err.Error(), html.WithStatus(http.StatusUnprocessableEntity))
		return
	}
	providers, err := h.VCSProviders.List(r.Context(), params.Organization)
	if err != nil {
		html.Error(r, w, err.Error())
		return
	}
	h.renderPage(
		h.templates.newView(newViewProps{
			organization: params.Organization,
			providers:    providers,
		}),
		"new module",
		w,
		r,
		withOrganization(params.Organization),
		withBreadcrumbs(
			helpers.Breadcrumb{Name: "Modules", Link: paths.Modules(params.Organization)},
			helpers.Breadcrumb{Name: "new"},
		),
	)
}

func (h *Handlers) connectModule(w http.ResponseWriter, r *http.Request) {
	var params struct {
		VCSProviderID resource.TfeID `schema:"vcs_provider_id,required"`
		// TODO: filters, public/private, etc
	}
	if err := decode.All(&params, r); err != nil {
		html.Error(r, w, err.Error(), html.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	provider, err := h.VCSProviders.Get(r.Context(), params.VCSProviderID)
	if err != nil {
		html.Error(r, w, err.Error())
		return
	}

	client, err := h.VCSProviders.Get(r.Context(), params.VCSProviderID)
	if err != nil {
		html.Error(r, w, err.Error())
		return
	}

	// Retrieve repos and filter according to required naming format
	// '<something>-<name>-<provider>'
	results, err := client.ListRepositories(r.Context(), vcs.ListRepositoriesOptions{
		PageSize: resource.MaxPageSize,
	})
	if err != nil {
		html.Error(r, w, err.Error())
		return
	}
	filtered := make([]vcs.Repo, 0, len(results))
	for _, res := range results {
		_, _, err := module.Repo(res).Split()
		if err == module.ErrInvalidModuleRepo {
			continue // skip repo
		} else if err != nil {
			html.Error(r, w, err.Error())
			return
		}
		filtered = append(filtered, res)
	}

	h.renderPage(
		h.templates.connect(connectProps{
			repos:    filtered,
			provider: provider,
		}),
		"new module",
		w,
		r,
		withOrganization(provider.Organization),
		withBreadcrumbs(
			helpers.Breadcrumb{Name: "Modules", Link: paths.Modules(provider.Organization)},
			helpers.Breadcrumb{Name: "new"},
		),
	)
}

func (h *Handlers) publishModule(w http.ResponseWriter, r *http.Request) {
	var params struct {
		VCSProviderID resource.TfeID `schema:"vcs_provider_id,required"`
		Repo          vcs.Repo       `schema:"identifier,required"`
	}
	if err := decode.All(&params, r); err != nil {
		html.Error(r, w, err.Error(), html.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	mod, err := h.Modules.PublishModule(r.Context(), module.PublishOptions{
		Repo:          module.Repo(params.Repo),
		VCSProviderID: params.VCSProviderID,
	})
	if errors.Is(err, vcs.ErrInvalidRepo) || errors.Is(err, module.ErrInvalidModuleRepo) {
		html.Error(r, w, err.Error(), html.WithStatus(http.StatusUnprocessableEntity))
		return
	} else if err != nil {
		html.Error(r, w, err.Error())
		return
	}

	html.FlashSuccess(w, "published module: "+mod.Name)
	http.Redirect(w, r, paths.Module(mod.ID), http.StatusFound)
}

func (h *Handlers) deleteModule(w http.ResponseWriter, r *http.Request) {
	id, err := decode.ID("module_id", r)
	if err != nil {
		html.Error(r, w, err.Error(), html.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	deleted, err := h.Modules.DeleteModule(r.Context(), id)
	if err != nil {
		html.Error(r, w, err.Error())
		return
	}

	html.FlashSuccess(w, "deleted module: "+deleted.Name)
	http.Redirect(w, r, paths.Modules(deleted.Organization), http.StatusFound)
}
