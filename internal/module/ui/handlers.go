package ui

import (
	"context"
	"errors"
	"html/template"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/module"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/ui/helpers"
	"github.com/leg100/otf/internal/ui/paths"
	"github.com/leg100/otf/internal/vcs"
)

type Handlers struct {
	client     Client
	authorizer authz.Interface
}

type Client interface {
	GetModuleByID(context.Context, resource.ID) (*module.Module, error)
	ListModules(context.Context, module.ListOptions) ([]*module.Module, error)
	ListProviders(context.Context, organization.Name) ([]string, error)
	GetModuleInfo(context.Context, resource.ID) (*module.TerraformModule, error)
	PublishModule(context.Context, module.PublishOptions) (*module.Module, error)
	DeleteModule(context.Context, resource.ID) (*module.Module, error)
	ListVCSProviders(ctx context.Context, organization organization.Name) ([]*vcs.Provider, error)
	GetVCSProvider(ctx context.Context, id resource.ID) (*vcs.Provider, error)
	Hostname() string
}

func NewHandlers(modules Client, authorizer authz.Interface) *Handlers {
	return &Handlers{
		client:     modules,
		authorizer: authorizer,
	}
}

func (h *Handlers) AddHandlers(r *mux.Router) {
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
		helpers.Error(r, w, err.Error(), helpers.WithStatus(http.StatusUnprocessableEntity))
		return
	}
	modules, err := h.client.ListModules(r.Context(), params.ListOptions)
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}
	providers, err := h.client.ListProviders(r.Context(), params.Organization)
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}

	props := moduleListProps{
		organization:          params.Organization,
		page:                  resource.NewPage(modules, params.PageOptions, nil),
		canPublishModule:      h.authorizer.CanAccess(r.Context(), authz.CreateModuleAction, params.Organization),
		providerFilterVisible: params.ProviderFilterVisible,
		allProviders:          providers,
		selectedProviders:     params.Providers,
	}
	helpers.RenderPage(
		moduleList(props),
		"Modules",
		w,
		r,
		helpers.WithOrganization(params.Organization),
		helpers.WithBreadcrumbs(
			helpers.Breadcrumb{Name: "Modules"},
		),
		helpers.WithContentActions(moduleListActions(props)),
	)
}

func (h *Handlers) getModule(w http.ResponseWriter, r *http.Request) {
	var params struct {
		ID      resource.TfeID `schema:"module_id,required"`
		Version *string        `schema:"version"`
	}
	if err := decode.All(&params, r); err != nil {
		helpers.Error(r, w, err.Error(), helpers.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	mod, err := h.client.GetModuleByID(r.Context(), params.ID)
	if err != nil {
		helpers.Error(r, w, err.Error())
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
		tfmod, err = h.client.GetModuleInfo(r.Context(), modver.ID)
		if err != nil {
			helpers.Error(r, w, err.Error())
			return
		}
	}

	switch mod.Status {
	case module.ModuleStatusSetupComplete:
		if tfmod != nil {
			readme = helpers.MarkdownToHTML(tfmod.GetReadme())
		}
	}

	helpers.RenderPage(
		moduleGet(moduleGetProps{
			module:          mod,
			terraformModule: tfmod,
			readme:          readme,
			currentVersion:  modver,
			hostname:        h.client.Hostname(),
		}),
		"modules",
		w,
		r,
		helpers.WithOrganization(mod.Organization),
		helpers.WithBreadcrumbs(
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
		helpers.Error(r, w, err.Error(), helpers.WithStatus(http.StatusUnprocessableEntity))
		return
	}
	providers, err := h.client.ListVCSProviders(r.Context(), params.Organization)
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}
	helpers.RenderPage(
		newView(newViewProps{
			organization: params.Organization,
			providers:    providers,
		}),
		"new module",
		w,
		r,
		helpers.WithOrganization(params.Organization),
		helpers.WithBreadcrumbs(
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
		helpers.Error(r, w, err.Error(), helpers.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	provider, err := h.client.GetVCSProvider(r.Context(), params.VCSProviderID)
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}

	client, err := h.client.GetVCSProvider(r.Context(), params.VCSProviderID)
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}

	// Retrieve repos and filter according to required naming format
	// '<something>-<name>-<provider>'
	results, err := client.ListRepositories(r.Context(), vcs.ListRepositoriesOptions{
		PageSize: resource.MaxPageSize,
	})
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}
	filtered := make([]vcs.Repo, 0, len(results))
	for _, res := range results {
		_, _, err := module.Repo(res).Split()
		if err == module.ErrInvalidModuleRepo {
			continue // skip repo
		} else if err != nil {
			helpers.Error(r, w, err.Error())
			return
		}
		filtered = append(filtered, res)
	}

	helpers.RenderPage(
		connect(connectProps{
			repos:    filtered,
			provider: provider,
		}),
		"new module",
		w,
		r,
		helpers.WithOrganization(provider.Organization),
		helpers.WithBreadcrumbs(
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
		helpers.Error(r, w, err.Error(), helpers.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	mod, err := h.client.PublishModule(r.Context(), module.PublishOptions{
		Repo:          module.Repo(params.Repo),
		VCSProviderID: params.VCSProviderID,
	})
	if errors.Is(err, vcs.ErrInvalidRepo) || errors.Is(err, module.ErrInvalidModuleRepo) {
		helpers.Error(r, w, err.Error(), helpers.WithStatus(http.StatusUnprocessableEntity))
		return
	} else if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}

	helpers.FlashSuccess(w, "published module: "+mod.Name)
	http.Redirect(w, r, paths.Module(mod.ID), http.StatusFound)
}

func (h *Handlers) deleteModule(w http.ResponseWriter, r *http.Request) {
	id, err := decode.ID("module_id", r)
	if err != nil {
		helpers.Error(r, w, err.Error(), helpers.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	deleted, err := h.client.DeleteModule(r.Context(), id)
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}

	helpers.FlashSuccess(w, "deleted module: "+deleted.Name)
	http.Redirect(w, r, paths.Modules(deleted.Organization), http.StatusFound)
}
