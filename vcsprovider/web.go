package vcsprovider

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/leg100/otf/cloud"
	"github.com/leg100/otf/http/decode"
	"github.com/leg100/otf/http/html"
	"github.com/leg100/otf/http/html/paths"
)

type webHandlers struct {
	otf.Renderer
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
	type parameters struct {
		Organization string `schema:"organization_name,required"`
		Cloud        string `schema:"cloud,required"`
	}
	var params parameters
	if err := decode.All(&params, r); err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	tmpl := fmt.Sprintf("vcs_provider_%s_new.tmpl", params.Cloud)
	h.Render(tmpl, w, r, params)
}

func (h *webHandlers) create(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		OrganizationName string `schema:"organization_name,required"`
		Token            string `schema:"token,required"`
		Name             string `schema:"name,required"`
		Cloud            string `schema:"cloud,required"`
	}
	var params parameters
	if err := decode.All(&params, r); err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	provider, err := h.svc.CreateVCSProvider(r.Context(), CreateOptions{
		Organization: params.OrganizationName,
		Token:        params.Token,
		Name:         params.Name,
		Cloud:        params.Cloud,
	})
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	html.FlashSuccess(w, "created provider: "+provider.Name)
	http.Redirect(w, r, paths.VCSProviders(provider.Organization), http.StatusFound)
}

func (h *webHandlers) list(w http.ResponseWriter, r *http.Request) {
	organization, err := decode.Param("organization_name", r)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	providers, err := h.svc.ListVCSProviders(r.Context(), organization)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.Render("vcs_provider_list.tmpl", w, r, struct {
		Items        []*VCSProvider
		CloudConfigs []cloud.Config
		Organization string
	}{
		Items:        providers,
		CloudConfigs: h.ListCloudConfigs(),
		Organization: organization,
	})
}

func (h *webHandlers) delete(w http.ResponseWriter, r *http.Request) {
	id, err := decode.Param("vcs_provider_id", r)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	provider, err := h.svc.DeleteVCSProvider(r.Context(), id)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	html.FlashSuccess(w, "deleted provider: "+provider.Name)
	http.Redirect(w, r, paths.VCSProviders(provider.Organization), http.StatusFound)
}
