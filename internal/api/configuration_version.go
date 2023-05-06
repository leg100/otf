package api

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/leg100/otf/api/types"
	"github.com/leg100/otf/configversion"
	otfhttp "github.com/leg100/otf/http"
	"github.com/leg100/otf/http/decode"
)

func (a *api) addConfigHandlers(r *mux.Router) {
	signed := r.PathPrefix("/signed/{signature.expiry}").Subrouter()
	signed.Use(otf.VerifySignedURL(a.Verifier))
	signed.HandleFunc("/configuration-versions/{id}/upload", a.uploadConfigurationVersion()).Methods("PUT")

	r = otfhttp.APIRouter(r)
	r.HandleFunc("/workspaces/{workspace_id}/configuration-versions", a.createConfigurationVersion).Methods("POST")
	r.HandleFunc("/configuration-versions/{id}", a.getConfigurationVersion).Methods("GET")
	r.HandleFunc("/workspaces/{workspace_id}/configuration-versions", a.listConfigurationVersions).Methods("GET")
	r.HandleFunc("/configuration-versions/{id}/download", a.downloadConfigurationVersion).Methods("GET")
}

func (a *api) createConfigurationVersion(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.Param("workspace_id", r)
	if err != nil {
		Error(w, err)
		return
	}

	opts := types.ConfigurationVersionCreateOptions{}
	if err := unmarshal(r.Body, &opts); err != nil {
		Error(w, err)
		return
	}
	cv, err := a.CreateConfigurationVersion(r.Context(), workspaceID, configversion.ConfigurationVersionCreateOptions{
		AutoQueueRuns: opts.AutoQueueRuns,
		Speculative:   opts.Speculative,
	})
	if err != nil {
		Error(w, err)
		return
	}

	a.writeResponse(w, r, cv, withCode(http.StatusCreated))
}

func (a *api) getConfigurationVersion(w http.ResponseWriter, r *http.Request) {
	id, err := decode.Param("id", r)
	if err != nil {
		Error(w, err)
		return
	}

	cv, err := a.GetConfigurationVersion(r.Context(), id)
	if err != nil {
		Error(w, err)
		return
	}
	a.writeResponse(w, r, cv)
}

func (a *api) listConfigurationVersions(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		WorkspaceID     string `schema:"workspace_id,required"`
		otf.ListOptions        // Pagination
	}
	var params parameters
	if err := decode.All(&params, r); err != nil {
		Error(w, err)
		return
	}

	cvl, err := a.ListConfigurationVersions(r.Context(), params.WorkspaceID, configversion.ConfigurationVersionListOptions{
		ListOptions: params.ListOptions,
	})
	if err != nil {
		Error(w, err)
		return
	}

	a.writeResponse(w, r, cvl)
}

func (a *api) uploadConfigurationVersion() http.HandlerFunc {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, err := decode.Param("id", r)
		if err != nil {
			Error(w, err)
			return
		}

		buf := new(bytes.Buffer)
		if _, err := io.Copy(buf, r.Body); err != nil {
			maxBytesError := &http.MaxBytesError{}
			if errors.As(err, &maxBytesError) {
				Error(w, &otf.HTTPError{
					Code:    422,
					Message: fmt.Sprintf("config exceeds maximum size (%d bytes)", a.maxConfigSize),
				})
			} else {
				Error(w, err)
			}
			return
		}
		if err := a.UploadConfig(r.Context(), id, buf.Bytes()); err != nil {
			Error(w, err)
			return
		}
	})
	return http.MaxBytesHandler(h, a.maxConfigSize).ServeHTTP
}

func (a *api) downloadConfigurationVersion(w http.ResponseWriter, r *http.Request) {
	id, err := decode.Param("id", r)
	if err != nil {
		Error(w, err)
		return
	}

	resp, err := a.DownloadConfig(r.Context(), id)
	if err != nil {
		Error(w, err)
		return
	}

	w.Write(resp)
}
