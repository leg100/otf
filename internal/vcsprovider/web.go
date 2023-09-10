package vcsprovider

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal/cloud"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/http/html"
	"github.com/leg100/otf/internal/http/html/paths"
	"github.com/leg100/otf/internal/organization"
)

type webHandlers struct {
	html.Renderer
	CloudService

	svc Service
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
		Organization string `schema:"organization_name,required"`
		Cloud        string `schema:"cloud,required"`
	}
	if err := decode.All(&params, r); err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	tmpl := fmt.Sprintf("vcs_provider_%s_new.tmpl", params.Cloud)
	h.Render(tmpl, w, struct {
		organization.OrganizationPage
		Cloud string
	}{
		OrganizationPage: organization.NewPage(r, "new vcs provider", params.Organization),
		Cloud:            params.Cloud,
	})
}

func (h *webHandlers) create(w http.ResponseWriter, r *http.Request) {
	var params struct {
		OrganizationName string `schema:"organization_name,required"`
		Token            string `schema:"token,required"`
		Name             string `schema:"name,required"`
		Cloud            string `schema:"cloud,required"`
	}
	if err := decode.All(&params, r); err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	provider, err := h.svc.CreateVCSProvider(r.Context(), CreateOptions{
		Organization: params.OrganizationName,
		Token:        params.Token,
		Name:         &params.Name,
		Cloud:        params.Cloud,
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
	}{
		OrganizationPage: organization.NewPage(r, "edit vcs provider", provider.Organization),
		VCSProvider:      provider,
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
		Name: &params.Name,
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

	providers, err := h.svc.ListVCSProviders(r.Context(), org)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.Render("vcs_provider_list.tmpl", w, struct {
		organization.OrganizationPage
		Items        []*VCSProvider
		CloudConfigs []cloud.Config
	}{
		OrganizationPage: organization.NewPage(r, "vcs providers", org),
		Items:            providers,
		CloudConfigs:     h.ListCloudConfigs(),
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
