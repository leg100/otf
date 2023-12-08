package module

import (
	"context"
	"errors"
	"html/template"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/http/html"
	"github.com/leg100/otf/internal/http/html/paths"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/rbac"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/vcs"
	"github.com/leg100/otf/internal/vcsprovider"
)

const (
	newModuleConnectStep newModuleStep = "connect-vcs"
	newModuleRepoStep    newModuleStep = "select-repo"
	newModuleConfirmStep newModuleStep = "confirm-selection"
)

type (
	// webHandlers provides handlers for the webUI
	webHandlers struct {
		internal.Signer
		html.Renderer

		client       webModulesClient
		vcsproviders vcsprovidersClient
		system       webHostnameClient
	}

	webHostnameClient interface {
		Hostname() string
	}

	// webModulesClient provides web handlers with access to modules
	webModulesClient interface {
		GetModuleByID(ctx context.Context, id string) (*Module, error)
		GetModuleInfo(ctx context.Context, versionID string) (*TerraformModule, error)
		ListModules(context.Context, ListModulesOptions) ([]*Module, error)
		PublishModule(context.Context, PublishOptions) (*Module, error)
		DeleteModule(ctx context.Context, id string) (*Module, error)
	}

	// vcsprovidersClient provides web handlers with access to vcs providers
	vcsprovidersClient interface {
		Get(context.Context, string) (*vcsprovider.VCSProvider, error)
		List(context.Context, string) ([]*vcsprovider.VCSProvider, error)
		GetVCSClient(ctx context.Context, providerID string) (vcs.Client, error)
	}

	newModuleStep string
)

func (h *webHandlers) addHandlers(r *mux.Router) {
	r = html.UIRouter(r)

	r.HandleFunc("/organizations/{organization_name}/modules", h.list).Methods("GET")
	r.HandleFunc("/organizations/{organization_name}/modules/new", h.new).Methods("GET")
	r.HandleFunc("/organizations/{organization_name}/modules/create", h.publish).Methods("POST")
	r.HandleFunc("/modules/{module_id}", h.get).Methods("GET")
	r.HandleFunc("/modules/{module_id}/delete", h.delete).Methods("POST")
}

func (h *webHandlers) list(w http.ResponseWriter, r *http.Request) {
	var opts ListModulesOptions
	if err := decode.All(&opts, r); err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	modules, err := h.client.ListModules(r.Context(), opts)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	subject, err := internal.SubjectFromContext(r.Context())
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.Render("module_list.tmpl", w, struct {
		organization.OrganizationPage
		Items            []*Module
		CanPublishModule bool
	}{
		OrganizationPage: organization.NewPage(r, "modules", opts.Organization),
		Items:            modules,
		CanPublishModule: subject.CanAccessOrganization(rbac.CreateModuleAction, opts.Organization),
	})
}

func (h *webHandlers) get(w http.ResponseWriter, r *http.Request) {
	var params struct {
		ID      string  `schema:"module_id,required"`
		Version *string `schema:"version"`
	}
	if err := decode.All(&params, r); err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	module, err := h.client.GetModuleByID(r.Context(), params.ID)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var (
		modver *ModuleVersion
		tfmod  *TerraformModule
		readme template.HTML
	)
	if params.Version != nil {
		modver = module.Version(*params.Version)
	} else {
		modver = module.Latest()
	}
	if modver != nil {
		tfmod, err = h.client.GetModuleInfo(r.Context(), modver.ID)
		if err != nil {
			h.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	switch module.Status {
	case ModuleStatusSetupComplete:
		if tfmod != nil {
			readme = html.MarkdownToHTML(tfmod.readme)
		}
	}

	h.Render("module_get.tmpl", w, struct {
		organization.OrganizationPage
		Module                    *Module
		TerraformModule           *TerraformModule
		Readme                    template.HTML
		CurrentVersion            *ModuleVersion
		Hostname                  string
		ModuleStatusPending       ModuleStatus
		ModuleStatusNoVersionTags ModuleStatus
		ModuleStatusSetupFailed   ModuleStatus
		ModuleStatusSetupComplete ModuleStatus
		ModuleVersionStatusOK     ModuleVersionStatus
	}{
		OrganizationPage:          organization.NewPage(r, module.ID, module.Organization),
		Module:                    module,
		TerraformModule:           tfmod,
		Readme:                    readme,
		CurrentVersion:            modver,
		Hostname:                  h.system.Hostname(),
		ModuleStatusPending:       ModuleStatusPending,
		ModuleStatusNoVersionTags: ModuleStatusNoVersionTags,
		ModuleStatusSetupFailed:   ModuleStatusSetupFailed,
		ModuleStatusSetupComplete: ModuleStatusSetupComplete,
		ModuleVersionStatusOK:     ModuleVersionStatusOK,
	})
}

func (h *webHandlers) new(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Step newModuleStep `schema:"step"`
	}
	if err := decode.All(&params, r); err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
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

func (h *webHandlers) newModuleConnect(w http.ResponseWriter, r *http.Request) {
	org, err := decode.Param("organization_name", r)
	if err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	providers, err := h.vcsproviders.List(r.Context(), org)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.Render("module_new.tmpl", w, struct {
		organization.OrganizationPage
		Items []*vcsprovider.VCSProvider
		Step  newModuleStep
	}{
		OrganizationPage: organization.NewPage(r, "new module", org),
		Items:            providers,
		Step:             newModuleConnectStep,
	})
}

func (h *webHandlers) newModuleRepo(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Organization  string `schema:"organization_name,required"`
		VCSProviderID string `schema:"vcs_provider_id,required"`
		// TODO: filters, public/private, etc
	}
	if err := decode.All(&params, r); err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	client, err := h.vcsproviders.GetVCSClient(r.Context(), params.VCSProviderID)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Retrieve repos and filter according to required naming format
	// '<something>-<name>-<provider>'
	results, err := client.ListRepositories(r.Context(), vcs.ListRepositoriesOptions{
		PageSize: resource.MaxPageSize,
	})
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	filtered := make([]string, 0, len(results))
	for _, res := range results {
		_, _, err := Repo(res).Split()
		if err == ErrInvalidModuleRepo {
			continue // skip repo
		} else if err != nil {
			h.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		filtered = append(filtered, res)
	}

	h.Render("module_new.tmpl", w, struct {
		organization.OrganizationPage
		Repos         []string
		VCSProviderID string
		Step          newModuleStep
	}{
		OrganizationPage: organization.NewPage(r, "new module", params.Organization),
		Repos:            filtered,
		VCSProviderID:    params.VCSProviderID,
		Step:             newModuleRepoStep,
	})
}

func (h *webHandlers) newModuleConfirm(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Organization  string `schema:"organization_name,required"`
		VCSProviderID string `schema:"vcs_provider_id,required"`
		Repo          string `schema:"identifier,required"`
	}
	if err := decode.All(&params, r); err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	vcsprov, err := h.vcsproviders.Get(r.Context(), params.VCSProviderID)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.Render("module_new.tmpl", w, struct {
		organization.OrganizationPage
		Step        newModuleStep
		Repo        string
		VCSProvider *vcsprovider.VCSProvider
	}{
		OrganizationPage: organization.NewPage(r, "new module", params.Organization),
		Step:             newModuleConfirmStep,
		Repo:             params.Repo,
		VCSProvider:      vcsprov,
	})
}

func (h *webHandlers) publish(w http.ResponseWriter, r *http.Request) {
	var params struct {
		VCSProviderID string `schema:"vcs_provider_id,required"`
		Repo          Repo   `schema:"identifier,required"`
	}
	if err := decode.All(&params, r); err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	module, err := h.client.PublishModule(r.Context(), PublishOptions{
		Repo:          params.Repo,
		VCSProviderID: params.VCSProviderID,
	})
	if err != nil && errors.Is(err, internal.ErrInvalidRepo) || errors.Is(err, ErrInvalidModuleRepo) {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	} else if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	html.FlashSuccess(w, "published module: "+module.Name)
	http.Redirect(w, r, paths.Module(module.ID), http.StatusFound)
}

func (h *webHandlers) delete(w http.ResponseWriter, r *http.Request) {
	id, err := decode.Param("module_id", r)
	if err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	deleted, err := h.client.DeleteModule(r.Context(), id)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	html.FlashSuccess(w, "deleted module: "+deleted.Name)
	http.Redirect(w, r, paths.Modules(deleted.Organization), http.StatusFound)
}
