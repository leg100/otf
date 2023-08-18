package configversion

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"time"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	otfhttp "github.com/leg100/otf/internal/http"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/tfeapi"
	"github.com/leg100/otf/internal/tfeapi/types"
	"github.com/leg100/surl"
)

type tfe struct {
	Service
	*surl.Signer
	*tfeapi.Responder

	maxConfigSize int64 // Maximum permitted config upload size in bytes
}

func (a *tfe) addHandlers(r *mux.Router) {
	signed := r.PathPrefix("/signed/{signature.expiry}").Subrouter()
	signed.Use(internal.VerifySignedURL(a.Signer))
	signed.HandleFunc("/configuration-versions/{id}/upload", a.uploadConfigurationVersion()).Methods("PUT")

	r = otfhttp.APIRouter(r)
	r.HandleFunc("/workspaces/{workspace_id}/configuration-versions", a.createConfigurationVersion).Methods("POST")
	r.HandleFunc("/configuration-versions/{id}", a.getConfigurationVersion).Methods("GET")
	r.HandleFunc("/workspaces/{workspace_id}/configuration-versions", a.listConfigurationVersions).Methods("GET")
	r.HandleFunc("/configuration-versions/{id}/download", a.downloadConfigurationVersion).Methods("GET")
}

func (a *tfe) createConfigurationVersion(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.Param("workspace_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	params := types.ConfigurationVersionCreateOptions{}
	if err := tfeapi.Unmarshal(r.Body, &params); err != nil {
		tfeapi.Error(w, err)
		return
	}

	opts := ConfigurationVersionCreateOptions{
		AutoQueueRuns: params.AutoQueueRuns,
		Speculative:   params.Speculative,
		Source:        SourceAPI,
	}
	if tfeapi.IsTerraformCLI(r) {
		opts.Source = SourceTerraform
	}

	cv, err := a.CreateConfigurationVersion(r.Context(), workspaceID, opts)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	// upload url is only provided in the response when *creating* configuration
	// version:
	//
	// https://developer.hashicorp.com/terraform/cloud-docs/api-docs/configuration-versions#configuration-files-upload-url
	uploadURL := fmt.Sprintf("/configuration-versions/%s/upload", cv.ID)
	uploadURL, err = a.Sign(uploadURL, time.Hour)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	// terraform CLI expects an absolute URL
	uploadURL = otfhttp.Absolute(r, uploadURL)

	a.Respond(w, r, a.convert(cv, uploadURL), http.StatusCreated)
}

func (a *tfe) getConfigurationVersion(w http.ResponseWriter, r *http.Request) {
	id, err := decode.Param("id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	cv, err := a.GetConfigurationVersion(r.Context(), id)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	a.Respond(w, r, a.convert(cv, ""), http.StatusOK)
}

func (a *tfe) listConfigurationVersions(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		WorkspaceID          string `schema:"workspace_id,required"`
		resource.PageOptions        // Pagination
	}
	var params parameters
	if err := decode.All(&params, r); err != nil {
		tfeapi.Error(w, err)
		return
	}

	page, err := a.ListConfigurationVersions(r.Context(), params.WorkspaceID, ConfigurationVersionListOptions{
		PageOptions: params.PageOptions,
	})
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	a.RespondWithPage(w, r, page.Items, page.Pagination)
}

func (a *tfe) uploadConfigurationVersion() http.HandlerFunc {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, err := decode.Param("id", r)
		if err != nil {
			tfeapi.Error(w, err)
			return
		}

		buf := new(bytes.Buffer)
		if _, err := io.Copy(buf, r.Body); err != nil {
			maxBytesError := &http.MaxBytesError{}
			if errors.As(err, &maxBytesError) {
				tfeapi.Error(w, &internal.HTTPError{
					Code:    422,
					Message: fmt.Sprintf("config exceeds maximum size (%d bytes)", a.maxConfigSize),
				})
			} else {
				tfeapi.Error(w, err)
			}
			return
		}
		if err := a.UploadConfig(r.Context(), id, buf.Bytes()); err != nil {
			tfeapi.Error(w, err)
			return
		}
	})
	return http.MaxBytesHandler(h, a.maxConfigSize).ServeHTTP
}

func (a *tfe) downloadConfigurationVersion(w http.ResponseWriter, r *http.Request) {
	id, err := decode.Param("id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	resp, err := a.DownloadConfig(r.Context(), id)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	w.Write(resp)
}

func (a *tfe) include(ctx context.Context, v any) ([]any, error) {
	dst := reflect.Indirect(reflect.ValueOf(v))

	// v must be a struct with a field named ConfigurationVersionID of kind string
	if dst.Kind() != reflect.Struct {
		return nil, nil
	}
	id := dst.FieldByName("ConfigurationVersionID")
	if !id.IsValid() {
		return nil, nil
	}
	if id.Kind() != reflect.String {
		return nil, nil
	}
	cv, err := a.GetConfigurationVersion(ctx, id.String())
	if err != nil {
		return nil, err
	}
	return []any{a.convert(cv, "")}, nil
}

func (a *tfe) includeIngressAttributes(ctx context.Context, v any) ([]any, error) {
	cv, ok := v.(*types.ConfigurationVersion)
	if !ok {
		return nil, nil
	}
	if cv.IngressAttributes == nil {
		return nil, nil
	}
	return []any{&types.IngressAttributes{
		ID:        internal.ConvertID(cv.ID, "ia"),
		CommitSHA: cv.IngressAttributes.CommitSHA,
		CommitURL: cv.IngressAttributes.CommitURL,
	}}, nil
}

func (a *tfe) convert(from *ConfigurationVersion, uploadURL string) *types.ConfigurationVersion {
	to := &types.ConfigurationVersion{
		ID:               from.ID,
		AutoQueueRuns:    from.AutoQueueRuns,
		Speculative:      from.Speculative,
		Source:           string(from.Source),
		Status:           string(from.Status),
		StatusTimestamps: &types.CVStatusTimestamps{},
		UploadURL:        uploadURL,
	}
	if from.IngressAttributes != nil {
		to.IngressAttributes = &types.IngressAttributes{
			ID: internal.ConvertID(from.ID, "ia"),
		}
	}
	for _, ts := range from.StatusTimestamps {
		switch ts.Status {
		case ConfigurationPending:
			to.StatusTimestamps.QueuedAt = &ts.Timestamp
		case ConfigurationErrored:
			to.StatusTimestamps.FinishedAt = &ts.Timestamp
		case ConfigurationUploaded:
			to.StatusTimestamps.StartedAt = &ts.Timestamp
		}
	}
	return to
}