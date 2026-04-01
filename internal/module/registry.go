package module

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/tfeapi"
)

type Registry struct {
	Client registryClient
	Signer tfeapi.Signer
}

type registryClient interface {
	GetModule(ctx context.Context, opts GetModuleOptions) (*Module, error)
	downloadVersion(ctx context.Context, versionID resource.ID) ([]byte, error)
}

// AddHandlers registers handlers for the module registry. It implements
// the Module Registry Protocol:
//
// https://developer.hashicorp.com/terraform/internals/module-registry-protocol
func (h *Registry) AddHandlers(r *mux.Router) {
	registry := r.PathPrefix(tfeapi.ModuleV1Prefix).Subrouter()
	registry.HandleFunc("/{organization}/{name}/{provider}/versions", h.listAvailableVersions).Methods("GET")
	registry.HandleFunc("/{organization}/{name}/{provider}/{version}/download", h.getModuleVersionDownloadLink).Methods("GET")

	// Signed paths
	signed := r.PathPrefix(tfeapi.SignedPrefixWithSignature).Subrouter()
	signed.HandleFunc("/modules/download/{module_version_id}.tar.gz", h.downloadModuleVersion).Methods("GET")
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
func (h *Registry) listAvailableVersions(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Name         string            `schema:"name,required"`
		Provider     string            `schema:"provider,required"`
		Organization organization.Name `schema:"organization,required"`
	}
	if err := decode.Route(&params, r); err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	mod, err := h.Client.GetModule(r.Context(), GetModuleOptions{
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

func (h *Registry) getModuleVersionDownloadLink(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Name         string            `schema:"name,required"`
		Provider     string            `schema:"provider,required"`
		Organization organization.Name `schema:"organization,required"`
		Version      string            `schema:"version,required"`
	}
	if err := decode.Route(&params, r); err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	mod, err := h.Client.GetModule(r.Context(), GetModuleOptions{
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

	signed, err := h.Signer.Sign(fmt.Sprintf("/modules/download/%s.tar.gz", version.ID), time.Now().Add(time.Hour))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Add("X-Terraform-Get", signed)
	w.WriteHeader(http.StatusNoContent)
}

func (h *Registry) downloadModuleVersion(w http.ResponseWriter, r *http.Request) {
	id, err := decode.ID("module_version_id", r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	tarball, err := h.Client.downloadVersion(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Write(tarball)
}
