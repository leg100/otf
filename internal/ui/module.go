package ui

import (
	"context"
	"errors"
	"html/template"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/http/html"
	"github.com/leg100/otf/internal/http/html/paths"
	"github.com/leg100/otf/internal/module"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/vcs"
)

type (
	// moduleHandlers provides handlers for the webUI
	moduleHandlers struct {
		client       moduleClient
		vcsproviders moduleVCSProvidersClient
		system       moduleHostnameClient
		authorizer   moduleAuthorizer
	}

	moduleHostnameClient interface {
		Hostname() string
	}

	// moduleClient provides web handlers with access to modules
	moduleClient interface {
		GetModuleByID(ctx context.Context, id resource.TfeID) (*module.Module, error)
		GetModuleInfo(ctx context.Context, versionID resource.TfeID) (*module.TerraformModule, error)
		ListModules(context.Context, module.ListOptions) ([]*module.Module, error)
		PublishModule(context.Context, module.PublishOptions) (*module.Module, error)
		DeleteModule(ctx context.Context, id resource.TfeID) (*module.Module, error)
		ListProviders(context.Context, organization.Name) ([]string, error)
	}

	moduleAuthorizer interface {
		CanAccess(context.Context, authz.Action, resource.ID) bool
	}

	// moduleVCSProvidersClient provides web handlers with access to vcs providers
	moduleVCSProvidersClient interface {
		Get(context.Context, resource.TfeID) (*vcs.Provider, error)
		List(context.Context, organization.Name) ([]*vcs.Provider, error)
	}
)

// AddModuleHandlers registers module UI handlers with the router
func AddModuleHandlers(r *mux.Router, client moduleClient, vcsproviders moduleVCSProvidersClient, system moduleHostnameClient, authorizer moduleAuthorizer) {
	h := &moduleHandlers{
		client:       client,
		vcsproviders: vcsproviders,
		system:       system,
		authorizer:   authorizer,
	}
	h.addHandlers(r)
}

func (h *moduleHandlers) addHandlers(r *mux.Router) {
	r = html.UIRouter(r)

	r.HandleFunc("/organizations/{organization_name}/modules", h.list).Methods("GET")
	r.HandleFunc("/organizations/{organization_name}/modules/new", h.new).Methods("GET")
	r.HandleFunc("/modules/{vcs_provider_id}/connect", h.connect).Methods("GET")
	r.HandleFunc("/organizations/{organization_name}/modules/create", h.publish).Methods("POST")
	r.HandleFunc("/modules/{module_id}", h.get).Methods("GET")
	r.HandleFunc("/modules/{module_id}/delete", h.delete).Methods("POST")
}

func (h *moduleHandlers) list(w http.ResponseWriter, r *http.Request) {
	var params struct {
		module.ListOptions
		resource.PageOptions
		ProviderFilterVisible bool `schema:"provider_filter_visible"`
	}
	if err := decode.All(&params, r); err != nil {
		html.Error(r, w, err.Error(), html.WithStatus(http.StatusUnprocessableEntity))
		return
	}
	modules, err := h.client.ListModules(r.Context(), params.ListOptions)
	if err != nil {
		html.Error(r, w, err.Error())
		return
	}
	providers, err := h.client.ListProviders(r.Context(), params.Organization)
	if err != nil {
		html.Error(r, w, err.Error())
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
	html.Render(moduleList(props), w, r)
}

func (h *moduleHandlers) get(w http.ResponseWriter, r *http.Request) {
	var params struct {
		ID      resource.TfeID `schema:"module_id,required"`
		Version *string        `schema:"version"`
	}
	if err := decode.All(&params, r); err != nil {
		html.Error(r, w, err.Error(), html.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	mod, err := h.client.GetModuleByID(r.Context(), params.ID)
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
		tfmod, err = h.client.GetModuleInfo(r.Context(), modver.ID)
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

	props := moduleGetProps{
		module:          mod,
		terraformModule: tfmod,
		readme:          readme,
		currentVersion:  modver,
		hostname:        h.system.Hostname(),
	}
	html.Render(moduleGet(props), w, r)
}

func (h *moduleHandlers) new(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Organization organization.Name `schema:"organization_name"`
	}
	if err := decode.All(&params, r); err != nil {
		html.Error(r, w, err.Error(), html.WithStatus(http.StatusUnprocessableEntity))
		return
	}
	providers, err := h.vcsproviders.List(r.Context(), params.Organization)
	if err != nil {
		html.Error(r, w, err.Error())
		return
	}
	props := newViewProps{
		organization: params.Organization,
		providers:    providers,
	}
	html.Render(newView(props), w, r)
}

func (h *moduleHandlers) connect(w http.ResponseWriter, r *http.Request) {
	var params struct {
		VCSProviderID resource.TfeID `schema:"vcs_provider_id,required"`
		// TODO: filters, public/private, etc
	}
	if err := decode.All(&params, r); err != nil {
		html.Error(r, w, err.Error(), html.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	provider, err := h.vcsproviders.Get(r.Context(), params.VCSProviderID)
	if err != nil {
		html.Error(r, w, err.Error())
		return
	}

	client, err := h.vcsproviders.Get(r.Context(), params.VCSProviderID)
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

	props := connectProps{
		repos:    filtered,
		provider: provider,
	}
	html.Render(connect(props), w, r)
}

func (h *moduleHandlers) publish(w http.ResponseWriter, r *http.Request) {
	var params struct {
		VCSProviderID resource.TfeID `schema:"vcs_provider_id,required"`
		Repo          vcs.Repo       `schema:"identifier,required"`
	}
	if err := decode.All(&params, r); err != nil {
		html.Error(r, w, err.Error(), html.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	mod, err := h.client.PublishModule(r.Context(), module.PublishOptions{
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

func (h *moduleHandlers) delete(w http.ResponseWriter, r *http.Request) {
	id, err := decode.ID("module_id", r)
	if err != nil {
		html.Error(r, w, err.Error(), html.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	deleted, err := h.client.DeleteModule(r.Context(), id)
	if err != nil {
		html.Error(r, w, err.Error())
		return
	}

	html.FlashSuccess(w, "deleted module: "+deleted.Name)
	http.Redirect(w, r, paths.Modules(deleted.Organization), http.StatusFound)
}
