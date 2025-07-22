package vcs

import (
	"context"
	"net/http"
	"strings"

	"github.com/a-h/templ"
	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/http/html"
	"github.com/leg100/otf/internal/http/html/paths"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/resource"
	"github.com/templ-go/x/urlbuilder"
)

type webHandlers struct {
	*internal.HostnameService

	client webClient
}

type webClient interface {
	Create(ctx context.Context, opts CreateOptions) (*Provider, error)
	Update(ctx context.Context, id resource.TfeID, opts UpdateOptions) (*Provider, error)
	Get(ctx context.Context, id resource.TfeID) (*Provider, error)
	List(ctx context.Context, organization organization.Name) ([]*Provider, error)
	Delete(ctx context.Context, id resource.TfeID) (*Provider, error)
	GetKind(id KindID) (Kind, error)
	GetKinds() []Kind
}

func (h *webHandlers) addHandlers(r *mux.Router) {
	r = html.UIRouter(r)

	r.HandleFunc("/organizations/{organization_name}/vcs-providers", h.list).Methods("GET")
	r.HandleFunc("/organizations/{organization_name}/vcs-providers/new", h.new).Methods("GET")
	r.HandleFunc("/organizations/{organization_name}/vcs-providers/create", h.create).Methods("POST")
	r.HandleFunc("/vcs-providers/{vcs_provider_id}/edit", h.edit).Methods("GET")
	r.HandleFunc("/vcs-providers/{vcs_provider_id}/update", h.update).Methods("POST")
	r.HandleFunc("/vcs-providers/{vcs_provider_id}/delete", h.delete).Methods("POST")
}

func (h *webHandlers) new(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Organization organization.Name `schema:"organization_name,required"`
		KindID       KindID            `schema:"kind,required"`
	}
	if err := decode.All(&params, r); err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	kind, err := h.client.GetKind(params.KindID)
	if err != nil {
		html.Error(w, "schema not found", http.StatusUnprocessableEntity)
		return
	}

	props := newProviderProps{
		organization: params.Organization,
		kind:         kind,
	}
	html.Render(newProvider(props), w, r)
}

func (h *webHandlers) create(w http.ResponseWriter, r *http.Request) {
	var params CreateOptions
	if err := decode.All(&params, r); err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	provider, err := h.client.Create(r.Context(), params)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	html.FlashSuccess(w, "created provider: "+provider.String())
	http.Redirect(w, r, paths.VCSProviders(provider.Organization), http.StatusFound)
}

func (h *webHandlers) edit(w http.ResponseWriter, r *http.Request) {
	providerID, err := decode.ID("vcs_provider_id", r)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	provider, err := h.client.Get(r.Context(), providerID)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	html.Render(edit(provider), w, r)
}

func (h *webHandlers) update(w http.ResponseWriter, r *http.Request) {
	var params struct {
		ID      resource.TfeID   `schema:"vcs_provider_id,required"`
		Token   string           `schema:"token"`
		Name    string           `schema:"name"`
		BaseURL *internal.WebURL `schema:"base_url,required"`
	}
	if err := decode.All(&params, r); err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	opts := UpdateOptions{
		Name:    params.Name,
		BaseURL: params.BaseURL,
	}
	// Because token is sensitive it's not sent to the browser, and so when this
	// handler is called, the token will be an empty string if user has not
	// updated it.
	if params.Token != "" {
		opts.Token = &params.Token
	}
	provider, err := h.client.Update(r.Context(), params.ID, opts)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	html.FlashSuccess(w, "updated provider: "+provider.String())
	http.Redirect(w, r, paths.VCSProviders(provider.Organization), http.StatusFound)
}

func (h *webHandlers) list(w http.ResponseWriter, r *http.Request) {
	var params ListOptions
	if err := decode.All(&params, r); err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	providers, err := h.client.List(r.Context(), params.Organization)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	props := listProps{
		organization: params.Organization,
		providers:    resource.NewPage(providers, params.PageOptions, nil),
		kinds:        h.client.GetKinds(),
	}
	html.Render(list(props), w, r)
}

func (h *webHandlers) delete(w http.ResponseWriter, r *http.Request) {
	id, err := decode.ID("vcs_provider_id", r)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	provider, err := h.client.Delete(r.Context(), id)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	html.FlashSuccess(w, "deleted provider: "+provider.String())
	http.Redirect(w, r, paths.VCSProviders(provider.Organization), http.StatusFound)
}

func RepoURL(provider *Provider, repo Repo) templ.SafeURL {
	b := urlbuilder.New(provider.BaseURL.Scheme, provider.BaseURL.Host)
	for segment := range strings.SplitSeq(repo.owner, "/") {
		b.Path(segment)
	}
	b.Path(repo.name)
	return b.Build()
}
