package module

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/tfeapi"
	"github.com/leg100/surl/v2"
)

type api struct {
	*surl.Signer

	svc *Service
}

func (h *api) addHandlers(r *mux.Router) {
	// signed routes
	signed := r.PathPrefix("/signed/{signature.expiry}").Subrouter()
	signed.Use(internal.VerifySignedURL(h.Signer))
	signed.HandleFunc("/modules/download/{module_version_id}.tar.gz", h.downloadModuleVersion).Methods("GET")

	// authenticated module api routes
	//
	// Implements the Module Registry Protocol:
	//
	// https://developer.hashicorp.com/terraform/internals/module-registry-protocol
	r = r.PathPrefix(tfeapi.ModuleV1Prefix).Subrouter()

	r.HandleFunc("/{organization}/{name}/{provider}/versions", h.listAvailableVersions).Methods("GET")
	r.HandleFunc("/{organization}/{name}/{provider}/{version}/download", h.getModuleVersionDownloadLink).Methods("GET")
}

type (
	listAvailableVersionsResponse struct {
		Modules []listAvailableVersionsModule
	}
	listAvailableVersionsModule struct {
		Source   string
		Versions []listAvailableVersionsVersion
	}
	listAvailableVersionsVersion struct {
		Version string
	}
)

// List Available Versions for a Specific Module.
//
// https://developer.hashicorp.com/terraform/registry/api-docs#list-available-versions-for-a-specific-module
func (h *api) listAvailableVersions(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Name         string                    `schema:"name,required"`
		Provider     string                    `schema:"provider,required"`
		Organization resource.OrganizationName `schema:"organization,required"`
	}
	if err := decode.Route(&params, r); err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	mod, err := h.svc.GetModule(r.Context(), GetModuleOptions{
		Name:         params.Name,
		Provider:     params.Provider,
		Organization: params.Organization,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-type", "application/json")

	response := listAvailableVersionsResponse{
		Modules: []listAvailableVersionsModule{
			{
				Source: strings.Join([]string{params.Organization.String(), params.Provider, params.Name}, "/"),
			},
		},
	}
	for _, ver := range mod.AvailableVersions() {
		response.Modules[0].Versions = append(response.Modules[0].Versions, listAvailableVersionsVersion{
			Version: ver.Version,
		})
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *api) getModuleVersionDownloadLink(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Name         string                    `schema:"name,required"`
		Provider     string                    `schema:"provider,required"`
		Organization resource.OrganizationName `schema:"organization,required"`
		Version      string                    `schema:"version,required"`
	}
	if err := decode.Route(&params, r); err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	mod, err := h.svc.GetModule(r.Context(), GetModuleOptions{
		Name:         params.Name,
		Provider:     params.Provider,
		Organization: params.Organization,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	version := mod.Version(params.Version)
	if version == nil {
		http.Error(w, "version not found", http.StatusNotFound)
		return
	}

	signed, err := h.Sign(fmt.Sprintf("/modules/download/%s.tar.gz", version.ID), time.Now().Add(time.Hour))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Add("X-Terraform-Get", signed)
	w.WriteHeader(http.StatusNoContent)
}

func (h *api) downloadModuleVersion(w http.ResponseWriter, r *http.Request) {
	id, err := decode.TfeID("module_version_id", r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	tarball, err := h.svc.downloadVersion(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Write(tarball)
}
