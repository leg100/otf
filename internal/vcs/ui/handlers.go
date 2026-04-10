package ui

import (
	"context"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/ui/helpers"
	"github.com/leg100/otf/internal/ui/paths"
	"github.com/leg100/otf/internal/vcs"
)

type Handlers struct {
	VCSProviders VCSService
}

type VCSService interface {
	GetKind(id vcs.KindID) (vcs.Kind, error)
	GetKinds() []vcs.Kind
	CreateVCSProvider(ctx context.Context, opts vcs.CreateOptions) (*vcs.Provider, error)
	GetVCSProvider(ctx context.Context, id resource.TfeID) (*vcs.Provider, error)
	UpdateVCSProvider(ctx context.Context, id resource.TfeID, opts vcs.UpdateOptions) (*vcs.Provider, error)
	ListVCSProviders(ctx context.Context, organization organization.Name) ([]*vcs.Provider, error)
	DeleteVCSProvider(ctx context.Context, id resource.TfeID) (*vcs.Provider, error)
}

func NewHandlers(vcsService VCSService) *Handlers {
	return &Handlers{
		VCSProviders: vcsService,
	}
}

func (h *Handlers) AddHandlers(r *mux.Router) {
	r.HandleFunc("/organizations/{organization_name}/vcs-providers", h.listVCSProviders).Methods("GET")
	r.HandleFunc("/organizations/{organization_name}/vcs-providers/new", h.newVCSProvider).Methods("GET")
	r.HandleFunc("/organizations/{organization_name}/vcs-providers/create", h.createVCSProvider).Methods("POST")
	r.HandleFunc("/vcs-providers/{vcs_provider_id}/edit", h.editVCSProvider).Methods("GET")
	r.HandleFunc("/vcs-providers/{vcs_provider_id}/update", h.updateVCSProvider).Methods("POST")
	r.HandleFunc("/vcs-providers/{vcs_provider_id}/delete", h.deleteVCSProvider).Methods("POST")
}

func (h *Handlers) newVCSProvider(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Organization organization.Name `schema:"organization_name,required"`
		KindID       vcs.KindID        `schema:"kind,required"`
	}
	if err := decode.All(&params, r); err != nil {
		helpers.Error(r, w, err.Error(), helpers.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	kind, err := h.VCSProviders.GetKind(params.KindID)
	if err != nil {
		helpers.Error(r, w, "schema not found", helpers.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	props := newProviderProps{
		organization: params.Organization,
		kind:         kind,
	}
	helpers.RenderPage(
		newProvider(props),
		"new vcs provider",
		w,
		r,
		helpers.WithOrganization(params.Organization),
		helpers.WithBreadcrumbs(helpers.Breadcrumb{Name: "New VCS Provider"}),
	)
}

func (h *Handlers) createVCSProvider(w http.ResponseWriter, r *http.Request) {
	var params vcs.CreateOptions
	if err := decode.All(&params, r); err != nil {
		helpers.Error(r, w, err.Error(), helpers.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	// Browsers convert \n to \r\n in the provider's token textarea input but
	// this can result in an invalid token for the user, so we undo this
	// conversion here.
	//
	// https://developer.mozilla.org/en-US/docs/Web/HTML/Reference/Elements/textarea#wrap
	if params.Token != nil {
		*params.Token = strings.ReplaceAll(*params.Token, "\r\n", "\n")
	}

	provider, err := h.VCSProviders.CreateVCSProvider(r.Context(), params)
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}
	helpers.FlashSuccess(w, "created provider: "+provider.String())
	http.Redirect(w, r, paths.VCSProviders(provider.Organization), http.StatusFound)
}

func (h *Handlers) editVCSProvider(w http.ResponseWriter, r *http.Request) {
	providerID, err := decode.ID("vcs_provider_id", r)
	if err != nil {
		helpers.Error(r, w, err.Error(), helpers.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	provider, err := h.VCSProviders.GetVCSProvider(r.Context(), providerID)
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}

	helpers.RenderPage(
		edit(provider),
		"edit vcs provider",
		w,
		r,
		helpers.WithOrganization(provider.Organization),
		helpers.WithBreadcrumbs(
			helpers.Breadcrumb{Name: "VCS Providers", Link: paths.VCSProviders(provider.Organization)},
			helpers.Breadcrumb{Name: provider.String()},
			helpers.Breadcrumb{Name: "settings"},
		),
	)
}

func (h *Handlers) updateVCSProvider(w http.ResponseWriter, r *http.Request) {
	var params struct {
		ID      resource.TfeID   `schema:"vcs_provider_id,required"`
		Token   string           `schema:"token"`
		Name    string           `schema:"name"`
		BaseURL *internal.WebURL `schema:"base_url,required"`
	}
	if err := decode.All(&params, r); err != nil {
		helpers.Error(r, w, err.Error(), helpers.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	opts := vcs.UpdateOptions{
		Name:    params.Name,
		BaseURL: params.BaseURL,
	}
	// Because token is sensitive it's not sent to the browser, and so when this
	// handler is called, the token will be an empty string if user has not
	// updated it.
	if params.Token != "" {
		opts.Token = &params.Token
	}
	provider, err := h.VCSProviders.UpdateVCSProvider(r.Context(), params.ID, opts)
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}
	helpers.FlashSuccess(w, "updated provider: "+provider.String())
	http.Redirect(w, r, paths.VCSProviders(provider.Organization), http.StatusFound)
}

func (h *Handlers) listVCSProviders(w http.ResponseWriter, r *http.Request) {
	var params vcs.ListOptions
	if err := decode.All(&params, r); err != nil {
		helpers.Error(r, w, err.Error(), helpers.WithStatus(http.StatusUnprocessableEntity))
		return
	}
	providers, err := h.VCSProviders.ListVCSProviders(r.Context(), params.Organization)
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}
	props := listProps{
		organization: params.Organization,
		providers:    resource.NewPage(providers, params.PageOptions, nil),
		kinds:        h.VCSProviders.GetKinds(),
	}
	helpers.RenderPage(
		list(props),
		"vcs providers",
		w,
		r,
		helpers.WithOrganization(params.Organization),
		helpers.WithBreadcrumbs(helpers.Breadcrumb{Name: "VCS Providers"}),
	)
}

func (h *Handlers) deleteVCSProvider(w http.ResponseWriter, r *http.Request) {
	id, err := decode.ID("vcs_provider_id", r)
	if err != nil {
		helpers.Error(r, w, err.Error(), helpers.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	provider, err := h.VCSProviders.DeleteVCSProvider(r.Context(), id)
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}
	helpers.FlashSuccess(w, "deleted provider: "+provider.String())
	http.Redirect(w, r, paths.VCSProviders(provider.Organization), http.StatusFound)
}
