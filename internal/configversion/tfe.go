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
	"github.com/leg100/otf/internal/configversion/source"
	otfhttp "github.com/leg100/otf/internal/http"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/tfeapi"
	"github.com/leg100/otf/internal/tfeapi/types"
	"github.com/leg100/surl/v2"
)

type tfe struct {
	tfeClient

	logr.Logger
	*surl.Signer
	*tfeapi.Responder

	maxConfigSize int64 // Maximum permitted config upload size in bytes
}

// tfeConfigsClient gives the tfe handlers access to config version services
type tfeClient interface {
	Create(ctx context.Context, workspaceid resource.TfeID, opts CreateOptions) (*ConfigurationVersion, error)
	Get(ctx context.Context, id resource.TfeID) (*ConfigurationVersion, error)
	GetLatest(ctx context.Context, workspaceID resource.TfeID) (*ConfigurationVersion, error)
	List(ctx context.Context, workspaceID resource.TfeID, opts ListOptions) (*resource.Page[*ConfigurationVersion], error)
	Delete(ctx context.Context, cvID resource.TfeID) error

	UploadConfig(ctx context.Context, id resource.TfeID, config []byte) error
	DownloadConfig(ctx context.Context, id resource.TfeID) ([]byte, error)
}

func (a *tfe) addHandlers(r *mux.Router) {
	signed := r.PathPrefix("/signed/{signature.expiry}").Subrouter()
	signed.Use(internal.VerifySignedURL(a.Signer))
	signed.HandleFunc("/configuration-versions/{id}/upload", a.uploadConfigurationVersion()).Methods("PUT")

	r = r.PathPrefix(tfeapi.APIPrefixV2).Subrouter()
	r.HandleFunc("/workspaces/{workspace_id}/configuration-versions", a.createConfigurationVersion).Methods("POST")
	r.HandleFunc("/configuration-versions/{id}", a.getConfigurationVersion).Methods("GET")
	r.HandleFunc("/workspaces/{workspace_id}/configuration-versions", a.listConfigurationVersions).Methods("GET")
	r.HandleFunc("/configuration-versions/{id}/download", a.downloadConfigurationVersion).Methods("GET")
}

func (a *tfe) createConfigurationVersion(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.ID("workspace_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	params := TFEConfigurationVersionCreateOptions{}
	if err := tfeapi.Unmarshal(r.Body, &params); err != nil {
		tfeapi.Error(w, err)
		return
	}

	opts := CreateOptions{
		AutoQueueRuns: params.AutoQueueRuns,
		Speculative:   params.Speculative,
		Source:        source.API,
	}
	if tfeapi.IsTerraformCLI(r) {
		opts.Source = source.Terraform
	}

	cv, err := a.Create(r.Context(), workspaceID, opts)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	// upload url is only provided in the response when *creating* configuration
	// version:
	//
	// https://developer.hashicorp.com/terraform/cloud-docs/api-docs/configuration-versions#configuration-files-upload-url
	uploadURL := fmt.Sprintf("/configuration-versions/%s/upload", cv.ID)
	uploadURL, err = a.Sign(uploadURL, time.Now().Add(time.Hour))
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	// terraform CLI expects an absolute URL
	uploadURL = otfhttp.Absolute(r, uploadURL)

	a.Respond(w, r, a.convert(cv, uploadURL), http.StatusCreated)
}

func (a *tfe) getConfigurationVersion(w http.ResponseWriter, r *http.Request) {
	id, err := decode.ID("id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	cv, err := a.Get(r.Context(), id)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	a.Respond(w, r, a.convert(cv, ""), http.StatusOK)
}

func (a *tfe) listConfigurationVersions(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		WorkspaceID resource.TfeID `schema:"workspace_id,required"`
		types.ListOptions
	}
	var params parameters
	if err := decode.All(&params, r); err != nil {
		tfeapi.Error(w, err)
		return
	}

	page, err := a.List(r.Context(), params.WorkspaceID, ListOptions{
		PageOptions: resource.PageOptions(params.ListOptions),
	})
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	// convert items
	items := make([]*TFEConfigurationVersion, len(page.Items))
	for i, from := range page.Items {
		items[i] = a.convert(from, "")
	}
	a.RespondWithPage(w, r, items, page.Pagination)
}

func (a *tfe) uploadConfigurationVersion() http.HandlerFunc {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, err := decode.ID("id", r)
		if err != nil {
			tfeapi.Error(w, err)
			return
		}

		buf := new(bytes.Buffer)
		if _, err := io.Copy(buf, r.Body); err != nil {
			maxBytesError := &http.MaxBytesError{}
			if errors.As(err, &maxBytesError) {
				tfeapi.Error(w, err, tfeapi.WithStatus(http.StatusUnprocessableEntity))
				// Terraform CLI only informs the user that an HTTP 422 response
				// was received, and they aren't informed that their config
				// exceeds the max size. To help them diagnose the cause, the
				// error is logged on the server side too.
				a.Error(err, "uploaded config exceeds max size", "bytes", a.maxConfigSize)
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
	id, err := decode.ID("id", r)
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

	// v must be a struct with a field named ConfigurationVersionID of kind
	// resource.TfeID
	if dst.Kind() != reflect.Struct {
		return nil, nil
	}
	id := dst.FieldByName("ConfigurationVersionID")
	if !id.IsValid() {
		return nil, nil
	}
	resourceID, ok := id.Interface().(resource.TfeID)
	if !ok {
		return nil, nil
	}
	cv, err := a.Get(ctx, resourceID)
	if err != nil {
		return nil, err
	}
	return []any{a.convert(cv, "")}, nil
}

func (a *tfe) includeIngressAttributes(ctx context.Context, v any) ([]any, error) {
	tfeCV, ok := v.(*TFEConfigurationVersion)
	if !ok {
		return nil, nil
	}
	if tfeCV.IngressAttributes == nil {
		return nil, nil
	}
	// the tfe CV does not by default include ingress attributes, whereas the
	// otf CV *does*, so we need to fetch it.
	cv, err := a.Get(ctx, tfeCV.ID)
	if err != nil {
		return nil, err
	}
	return []any{&TFEIngressAttributes{
		ID:        resource.ConvertTfeID(cv.ID, resource.IngressAttributesKind),
		CommitSHA: cv.IngressAttributes.CommitSHA,
		CommitURL: cv.IngressAttributes.CommitURL,
	}}, nil
}

func (a *tfe) convert(from *ConfigurationVersion, uploadURL string) *TFEConfigurationVersion {
	to := &TFEConfigurationVersion{
		ID:               from.ID,
		AutoQueueRuns:    from.AutoQueueRuns,
		Speculative:      from.Speculative,
		Source:           string(from.Source),
		Status:           string(from.Status),
		StatusTimestamps: &TFECVStatusTimestamps{},
		UploadURL:        uploadURL,
	}
	if from.IngressAttributes != nil {
		to.IngressAttributes = &TFEIngressAttributes{
			ID: resource.ConvertTfeID(from.ID, "ia"),
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
