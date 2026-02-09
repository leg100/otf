package ui

import (
	"net/http"
	"strings"

	"github.com/a-h/templ"
	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/http/html"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/ui/helpers"
	"github.com/leg100/otf/internal/ui/paths"
	"github.com/leg100/otf/internal/vcs"
	"github.com/templ-go/x/urlbuilder"
)

func addVCSHandlers(r *mux.Router, h *Handlers) {
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
		html.Error(r, w, err.Error(), html.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	kind, err := h.VCSProviders.GetKind(params.KindID)
	if err != nil {
		html.Error(r, w, "schema not found", html.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	props := newProviderProps{
		organization: params.Organization,
		kind:         kind,
	}
	h.renderPage(
		h.templates.newProvider(props),
		"new vcs provider",
		w,
		r,
		withOrganization(params.Organization),
		withBreadcrumbs(helpers.Breadcrumb{Name: "New VCS Provider"}),
	)
}

func (h *Handlers) createVCSProvider(w http.ResponseWriter, r *http.Request) {
	var params vcs.CreateOptions
	if err := decode.All(&params, r); err != nil {
		html.Error(r, w, err.Error(), html.WithStatus(http.StatusUnprocessableEntity))
		return
	}
	provider, err := h.VCSProviders.Create(r.Context(), params)
	if err != nil {
		html.Error(r, w, err.Error())
		return
	}
	html.FlashSuccess(w, "created provider: "+provider.String())
	http.Redirect(w, r, paths.VCSProviders(provider.Organization), http.StatusFound)
}

func (h *Handlers) editVCSProvider(w http.ResponseWriter, r *http.Request) {
	providerID, err := decode.ID("vcs_provider_id", r)
	if err != nil {
		html.Error(r, w, err.Error(), html.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	provider, err := h.VCSProviders.Get(r.Context(), providerID)
	if err != nil {
		html.Error(r, w, err.Error())
		return
	}

	h.renderPage(
		h.templates.edit(provider),
		"edit vcs provider",
		w,
		r,
		withOrganization(provider.Organization),
		withBreadcrumbs(
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
		html.Error(r, w, err.Error(), html.WithStatus(http.StatusUnprocessableEntity))
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
	provider, err := h.VCSProviders.Update(r.Context(), params.ID, opts)
	if err != nil {
		html.Error(r, w, err.Error())
		return
	}
	html.FlashSuccess(w, "updated provider: "+provider.String())
	http.Redirect(w, r, paths.VCSProviders(provider.Organization), http.StatusFound)
}

func (h *Handlers) listVCSProviders(w http.ResponseWriter, r *http.Request) {
	var params vcs.ListOptions
	if err := decode.All(&params, r); err != nil {
		html.Error(r, w, err.Error(), html.WithStatus(http.StatusUnprocessableEntity))
		return
	}
	providers, err := h.VCSProviders.List(r.Context(), params.Organization)
	if err != nil {
		html.Error(r, w, err.Error())
		return
	}
	props := listProps{
		organization: params.Organization,
		providers:    resource.NewPage(providers, params.PageOptions, nil),
		kinds:        h.VCSProviders.GetKinds(),
	}
	h.renderPage(
		h.templates.list(props),
		"vcs providers",
		w,
		r,
		withOrganization(params.Organization),
		withBreadcrumbs(helpers.Breadcrumb{Name: "VCS Providers"}),
	)
}

func (h *Handlers) deleteVCSProvider(w http.ResponseWriter, r *http.Request) {
	id, err := decode.ID("vcs_provider_id", r)
	if err != nil {
		html.Error(r, w, err.Error(), html.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	provider, err := h.VCSProviders.Delete(r.Context(), id)
	if err != nil {
		html.Error(r, w, err.Error())
		return
	}
	html.FlashSuccess(w, "deleted provider: "+provider.String())
	http.Redirect(w, r, paths.VCSProviders(provider.Organization), http.StatusFound)
}

func RepoURL(provider *vcs.Provider, repo vcs.Repo) templ.SafeURL {
	b := urlbuilder.New(provider.BaseURL.Scheme, provider.BaseURL.Host)
	for segment := range strings.SplitSeq(repo.Owner(), "/") {
		b.Path(segment)
	}
	b.Path(repo.Name())
	return b.Build()
}
