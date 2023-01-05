package http

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/leg100/otf"
	"github.com/leg100/otf/http/decode"
)

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

func (s *Server) listModuleVersions(w http.ResponseWriter, r *http.Request) {
	var opts otf.GetModuleOptions
	if err := decode.Route(&opts, r); err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	mod, err := s.GetModule(r.Context(), opts)
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

func (s *Server) getModuleVersionDownloadLink(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		otf.GetModuleOptions
		Version string
	}
	var params parameters
	if err := decode.Route(&params, r); err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	mod, err := s.GetModule(r.Context(), params.GetModuleOptions)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	version := mod.Version(params.Version)
	if version == nil {
		http.Error(w, "version not found", http.StatusNotFound)
		return
	}

	signed, err := s.Sign("/modules/download/"+version.ID()+".tar.gz", time.Hour)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Add("X-Terraform-Get", signed)
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) downloadModuleVersion(w http.ResponseWriter, r *http.Request) {
	id, err := decode.Param("module_version_id", r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	tarball, err := s.DownloadModuleVersion(r.Context(), otf.DownloadModuleOptions{
		ModuleVersionID: id,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Write(tarball)
}
