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

	r.HandleFunc("/organizations/{organization_name}/vcs-providers", h.list)
	r.HandleFunc("/organizations/{organization_name}/vcs-providers/new", h.new)
	r.HandleFunc("/organizations/{organization_name}/vcs-providers/create", h.create)
	r.HandleFunc("/vcs-providers/{vcs_provider_id}/delete", h.delete)
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
		Name:         params.Name,
		Cloud:        params.Cloud,
	})
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	html.FlashSuccess(w, "created provider: "+provider.Name)
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
	html.FlashSuccess(w, "deleted provider: "+provider.Name)
	http.Redirect(w, r, paths.VCSProviders(provider.Organization), http.StatusFound)
}
