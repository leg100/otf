package module

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	otfhttp "github.com/leg100/otf/http"
	"github.com/leg100/otf/http/decode"
	"github.com/leg100/surl"
)

type handlers struct {
	*surl.Signer

	app appService
}

// Implements the Module Registry Protocol:
//
// https://developer.hashicorp.com/terraform/internals/module-registry-protocol
func (h *handlers) AddHandlers(r *mux.Router) {
	r.HandleFunc("/{organization}/{name}/{provider}/versions", h.listModuleVersions)
	r.HandleFunc("/{organization}/{name}/{provider}/{version}/download", h.getModuleVersionDownloadLink)

	signed := r.PathPrefix("/signed/{signature.expiry}").Subrouter()
	signed.Use((&otfhttp.SignatureVerifier{h.Signer}).Handler)
	signed.HandleFunc("/modules/download/{module_version_id}.tar.gz", h.downloadModuleVersion).Methods("GET")
}

type listModuleVersionsResponse struct {
	Modules []module
}

type module struct {
	Source   string
	Versions []moduleVersion
}

type moduleVersion struct {
	Version string
}

func (h *handlers) listModuleVersions(w http.ResponseWriter, r *http.Request) {
	var opts GetModuleOptions
	if err := decode.Route(&opts, r); err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	mod, err := h.app.GetModule(r.Context(), opts)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-type", "application/json")
	response := listModuleVersionsResponse{
		Modules: []module{
			{
				Source: strings.Join([]string{opts.Organization, opts.Provider, opts.Name}, "/"),
			},
		},
	}
	for _, ver := range mod.Versions() {
		response.Modules[0].Versions = append(response.Modules[0].Versions, moduleVersion{
			Version: ver.Version(),
		})
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *handlers) getModuleVersionDownloadLink(w http.ResponseWriter, r *http.Request) {
	params := struct {
		GetModuleOptions
		Version string
	}{}
	if err := decode.Route(&params, r); err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	mod, err := h.app.GetModule(r.Context(), params.GetModuleOptions)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	version := mod.Version(params.Version)
	if version == nil {
		http.Error(w, "version not found", http.StatusNotFound)
		return
	}

	signed, err := h.Sign("/modules/download/"+version.ID()+".tar.gz", time.Hour)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Add("X-Terraform-Get", signed)
	w.WriteHeader(http.StatusNoContent)
}

func (h *handlers) downloadModuleVersion(w http.ResponseWriter, r *http.Request) {
	id, err := decode.Param("module_version_id", r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	tarball, err := h.app.DownloadModuleVersion(r.Context(), DownloadModuleOptions{
		ModuleVersionID: id,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Write(tarball)
}
