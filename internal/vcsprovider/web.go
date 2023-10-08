package vcsprovider

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/github"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/http/html"
	"github.com/leg100/otf/internal/http/html/paths"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/vcs"
)

type webHandlers struct {
	html.Renderer
	internal.HostnameService
	github.GithubAppService

	svc            Service
	GithubHostname string
	GitlabHostname string
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
	r.HandleFunc("/vcs-providers/{vcs_provider_id}", h.get).Methods("GET")
}

func (h *webHandlers) newPersonalToken(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Organization string   `schema:"organization_name,required"`
		Kind         vcs.Kind `schema:"kind,required"`
	}
	if err := decode.All(&params, r); err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	response := struct {
		organization.OrganizationPage
		VCSProvider *VCSProvider
		FormAction  string
		EditMode    bool
		TokensURL   string
		Scope       string
		Kind        string
	}{
		OrganizationPage: organization.NewPage(r, "new vcs provider", params.Organization),
		VCSProvider:      &VCSProvider{Kind: params.Kind},
		FormAction:       paths.CreateVCSProvider(params.Organization),
		EditMode:         false,
	}
	switch params.Kind {
	case vcs.GithubKind:
		response.Kind = string(vcs.GithubKind)
		response.Scope = "repo"
		response.TokensURL = "https://" + h.GithubHostname + "/settings/tokens"
	case vcs.GitlabKind:
		response.Kind = string(vcs.GitlabKind)
		response.Scope = "api"
		response.TokensURL = "https://" + h.GitlabHostname + "/-/profile/personal_access_tokens"
	}
	h.Render("vcs_provider_pat_new.tmpl", w, response)
}

func (h *webHandlers) newGithubApp(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Organization string `schema:"organization_name,required"`
	}
	if err := decode.All(&params, r); err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	app, err := h.GetGithubApp(r.Context())
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	installs, err := h.ListInstallations(r.Context())
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.Render("vcs_provider_github_app_new.tmpl", w, struct {
		organization.OrganizationPage
		App            *github.App
		Installations  []*github.Installation
		Kind           vcs.Kind
		GithubHostname string
	}{
		OrganizationPage: organization.NewPage(r, "new vcs provider", params.Organization),
		App:              app,
		Installations:    installs,
		Kind:             vcs.GithubKind,
		GithubHostname:   h.GithubHostname,
	})
}

func (h *webHandlers) create(w http.ResponseWriter, r *http.Request) {
	var params struct {
		OrganizationName   string    `schema:"organization_name,required"`
		Token              *string   `schema:"token"`
		GithubAppInstallID *int64    `schema:"install_id"`
		Name               string    `schema:"name"`
		Kind               *vcs.Kind `schema:"kind"`
	}
	if err := decode.All(&params, r); err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	provider, err := h.svc.CreateVCSProvider(r.Context(), CreateOptions{
		Organization:       params.OrganizationName,
		Token:              params.Token,
		GithubAppInstallID: params.GithubAppInstallID,
		Name:               params.Name,
		Kind:               params.Kind,
	})
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	html.FlashSuccess(w, "created provider: "+provider.String())
	http.Redirect(w, r, paths.VCSProviders(provider.Organization), http.StatusFound)
}

func (h *webHandlers) edit(w http.ResponseWriter, r *http.Request) {
	providerID, err := decode.Param("vcs_provider_id", r)
	if err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	provider, err := h.svc.GetVCSProvider(r.Context(), providerID)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.Render("vcs_provider_edit.tmpl", w, struct {
		organization.OrganizationPage
		VCSProvider *VCSProvider
		FormAction  string
		EditMode    bool
	}{
		OrganizationPage: organization.NewPage(r, "edit vcs provider", provider.Organization),
		VCSProvider:      provider,
		FormAction:       paths.UpdateVCSProvider(providerID),
		EditMode:         true,
	})
}

func (h *webHandlers) update(w http.ResponseWriter, r *http.Request) {
	var params struct {
		ID    string `schema:"vcs_provider_id,required"`
		Token string `schema:"token"`
		Name  string `schema:"name"`
	}
	if err := decode.All(&params, r); err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	opts := UpdateOptions{
		Name: params.Name,
	}
	// avoid setting token to empty string
	if params.Token != "" {
		opts.Token = &params.Token
	}
	provider, err := h.svc.UpdateVCSProvider(r.Context(), params.ID, opts)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	html.FlashSuccess(w, "updated provider: "+provider.String())
	http.Redirect(w, r, paths.VCSProviders(provider.Organization), http.StatusFound)
}

func (h *webHandlers) list(w http.ResponseWriter, r *http.Request) {
	org, err := decode.Param("organization_name", r)
	if err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	app, err := h.GetGithubApp(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	providers, err := h.svc.ListVCSProviders(r.Context(), org)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.Render("vcs_provider_list.tmpl", w, struct {
		organization.OrganizationPage
		Items     []*VCSProvider
		GithubApp *github.App
	}{
		OrganizationPage: organization.NewPage(r, "vcs providers", org),
		Items:            providers,
		GithubApp:        app,
	})
}

func (h *webHandlers) get(w http.ResponseWriter, r *http.Request) {
	id, err := decode.Param("vcs_provider_id", r)
	if err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	provider, err := h.svc.GetVCSProvider(r.Context(), id)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.Render("vcs_provider_get.tmpl", w, struct {
		VCSProvider *VCSProvider
	}{
		VCSProvider: provider,
	})
}

func (h *webHandlers) delete(w http.ResponseWriter, r *http.Request) {
	id, err := decode.Param("vcs_provider_id", r)
	if err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	provider, err := h.svc.DeleteVCSProvider(r.Context(), id)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	html.FlashSuccess(w, "deleted provider: "+provider.String())
	http.Redirect(w, r, paths.VCSProviders(provider.Organization), http.StatusFound)
}
