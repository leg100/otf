package module

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/leg100/otf/http/decode"
	"github.com/leg100/surl"
)

type api struct {
	*surl.Signer

	svc service
}

// Implements the Module Registry Protocol:
//
// https://developer.hashicorp.com/terraform/internals/module-registry-protocol
func (h *api) addHandlers(r *mux.Router) {
	r.HandleFunc("/{organization}/{name}/{provider}/versions", h.listModuleVersions)
	r.HandleFunc("/{organization}/{name}/{provider}/{version}/download", h.getModuleVersionDownloadLink)

	signed := r.PathPrefix("/signed/{signature.expiry}").Subrouter()
	signed.Use(otf.VerifySignedURL(h.Signer))
	signed.HandleFunc("/modules/download/{module_version_id}.tar.gz", h.downloadModuleVersion).Methods("GET")
}

type module struct {
	Source   string
	Versions []moduleVersion
}

type moduleVersion struct {
	Version string
}

func (h *api) listModuleVersions(w http.ResponseWriter, r *http.Request) {
	var opts GetModuleOptions
	if err := decode.Route(&opts, r); err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	mod, err := h.svc.GetModule(r.Context(), opts)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-type", "application/json")

	type responseVersion struct {
		Version string
	}
	type responseModule struct {
		Source   string
		Versions []responseVersion
	}
	response := struct {
		Modules []responseModule
	}{
		Modules: []responseModule{
			{
				Source: strings.Join([]string{opts.Organization, opts.Provider, opts.Name}, "/"),
			},
		},
	}
	for _, ver := range mod.versions {
		response.Modules[0].Versions = append(response.Modules[0].Versions, responseVersion{
			Version: ver.version,
		})
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *api) getModuleVersionDownloadLink(w http.ResponseWriter, r *http.Request) {
	var params struct {
		GetModuleOptions
		Version string
	}
	if err := decode.Route(&params, r); err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	mod, err := h.svc.GetModule(r.Context(), params.GetModuleOptions)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	version := mod.versions[params.Version]
	if version == nil {
		http.Error(w, "version not found", http.StatusNotFound)
		return
	}

	signed, err := h.Sign("/modules/download/"+version.id+".tar.gz", time.Hour)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Add("X-Terraform-Get", signed)
	w.WriteHeader(http.StatusNoContent)
}

func (h *api) downloadModuleVersion(w http.ResponseWriter, r *http.Request) {
	id, err := decode.Param("module_version_id", r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	tarball, err := h.svc.DownloadModuleVersion(r.Context(), DownloadModuleOptions{
		ModuleVersionID: id,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Write(tarball)
}
