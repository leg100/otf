package vcsprovider

import (
	"context"
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/github"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/http/html"
	"github.com/leg100/otf/internal/http/html/paths"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/vcs"
	"github.com/templ-go/x/urlbuilder"
)

type webHandlers struct {
	*internal.HostnameService

	client     webClient
	githubApps webGithubAppClient

	GithubHostname string
	GitlabHostname string
}

type webClient interface {
	Create(ctx context.Context, opts CreateOptions) (*VCSProvider, error)
	Update(ctx context.Context, id resource.TfeID, opts UpdateOptions) (*VCSProvider, error)
	Get(ctx context.Context, id resource.TfeID) (*VCSProvider, error)
	List(ctx context.Context, organization organization.Name) ([]*VCSProvider, error)
	Delete(ctx context.Context, id resource.TfeID) (*VCSProvider, error)
}

type webGithubAppClient interface {
	GetApp(ctx context.Context) (*github.App, error)
	ListInstallations(ctx context.Context) ([]*github.Installation, error)
}

func (h *webHandlers) addHandlers(r *mux.Router) {
	r = html.UIRouter(r)

	r.HandleFunc("/organizations/{organization_name}/vcs-providers", h.list).Methods("GET")
	r.HandleFunc("/organizations/{organization_name}/vcs-providers/new", h.newPersonalToken).Methods("GET")
	r.HandleFunc("/organizations/{organization_name}/vcs-providers/new-github-app", h.newGithubApp).Methods("GET")
	r.HandleFunc("/organizations/{organization_name}/vcs-providers/create", h.create).Methods("POST")
	r.HandleFunc("/vcs-providers/{vcs_provider_id}/edit", h.edit).Methods("GET")
	r.HandleFunc("/vcs-providers/{vcs_provider_id}/update", h.update).Methods("POST")
	r.HandleFunc("/vcs-providers/{vcs_provider_id}/delete", h.delete).Methods("POST")
}

func (h *webHandlers) newPersonalToken(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Organization organization.Name `schema:"organization_name,required"`
		Kind         vcs.Kind          `schema:"kind,required"`
	}
	if err := decode.All(&params, r); err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	props := newPATProps{
		provider: &VCSProvider{
			Kind:         params.Kind,
			Organization: params.Organization,
		},
	}
	switch params.Kind {
	case vcs.GithubKind:
		props.scope = "repo"
		props.tokensURL = urlbuilder.New("https", h.GithubHostname).Path("/settings/tokens").Build()
	case vcs.GitlabKind:
		props.scope = "api"
		props.tokensURL = urlbuilder.New("https", h.GitlabHostname).Path("/-/profile/personal_access_tokens").Build()
	}
	html.Render(newPAT(props), w, r)
}

func (h *webHandlers) newGithubApp(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Organization organization.Name `schema:"organization_name,required"`
	}
	if err := decode.All(&params, r); err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	app, err := h.githubApps.GetApp(r.Context())
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	installs, err := h.githubApps.ListInstallations(r.Context())
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	props := newGithubAppProps{
		organization:   params.Organization,
		app:            app,
		installations:  installs,
		kind:           vcs.GithubKind,
		githubHostname: h.GithubHostname,
	}
	html.Render(newGithubApp(props), w, r)
}

func (h *webHandlers) create(w http.ResponseWriter, r *http.Request) {
	var params struct {
		OrganizationName   organization.Name `schema:"organization_name,required"`
		Token              *string           `schema:"token"`
		GithubAppInstallID *int64            `schema:"install_id"`
		Name               string            `schema:"name"`
		Kind               *vcs.Kind         `schema:"kind"`
	}
	if err := decode.All(&params, r); err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	provider, err := h.client.Create(r.Context(), CreateOptions{
		Organization:       params.OrganizationName,
		Token:              params.Token,
		GithubAppInstallID: params.GithubAppInstallID,
		Name:               params.Name,
		Kind:               params.Kind,
	})
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
		ID    resource.TfeID `schema:"vcs_provider_id,required"`
		Token string         `schema:"token"`
		Name  string         `schema:"name"`
	}
	if err := decode.All(&params, r); err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	opts := UpdateOptions{
		Name: params.Name,
	}
	// avoid setting token to empty string
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
	app, err := h.githubApps.GetApp(r.Context())
	if errors.Is(err, internal.ErrResourceNotFound) {
		// app not found, which is ok.
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
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
		app:          app,
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
