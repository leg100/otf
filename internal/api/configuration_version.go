package api

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/DataDog/jsonapi"
	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/api/types"
	"github.com/leg100/otf/internal/configversion"
	otfhttp "github.com/leg100/otf/internal/http"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/resource"
)

func (a *api) addConfigHandlers(r *mux.Router) {
	signed := r.PathPrefix("/signed/{signature.expiry}").Subrouter()
	signed.Use(internal.VerifySignedURL(a.Signer))
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
	params := types.ConfigurationVersionCreateOptions{}
	if err := unmarshal(r.Body, &params); err != nil {
		Error(w, err)
		return
	}

	opts := configversion.ConfigurationVersionCreateOptions{
		AutoQueueRuns: params.AutoQueueRuns,
		Speculative:   params.Speculative,
		Source:        configversion.SourceAPI,
	}
	if r.Header.Get(headerSourceKey) == headerSourceValue {
		opts.Source = configversion.SourceTerraform
	}

	cv, err := a.CreateConfigurationVersion(r.Context(), workspaceID, opts)
	if err != nil {
		Error(w, err)
		return
	}

	to, _ := a.toConfigurationVersion(cv, r)

	// upload url is only provided in the response when creating configuration
	// version:
	//
	// https://developer.hashicorp.com/terraform/cloud-docs/api-docs/configuration-versions#configuration-files-upload-url
	uploadURL := fmt.Sprintf("/configuration-versions/%s/upload", cv.ID)
	uploadURL, err = a.Sign(uploadURL, time.Hour)
	if err != nil {
		Error(w, err)
		return
	}
	// terraform CLI expects an absolute URL
	to.UploadURL = otfhttp.Absolute(r, uploadURL)

	b, err := jsonapi.Marshal(to)
	if err != nil {
		Error(w, err)
		return
	}

	w.Header().Set("Content-type", mediaType)
	w.WriteHeader(http.StatusCreated)
	w.Write(b)
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
		WorkspaceID          string `schema:"workspace_id,required"`
		resource.PageOptions        // Pagination
	}
	var params parameters
	if err := decode.All(&params, r); err != nil {
		Error(w, err)
		return
	}

	cvl, err := a.ListConfigurationVersions(r.Context(), params.WorkspaceID, configversion.ConfigurationVersionListOptions{
		PageOptions: params.PageOptions,
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
				Error(w, &internal.HTTPError{
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
