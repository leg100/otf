package run

import (
	"bytes"
	"context"
	"io"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/tfeapi"
)

type api struct {
	*tfeapi.Responder
	Client apiClient
	Logger logr.Logger
}

type apiClient interface {
	CreateRun(context.Context, resource.TfeID, CreateOptions) (*Run, error)
	ListRuns(_ context.Context, opts ListOptions) (*resource.Page[*Run], error)
	GetRun(ctx context.Context, id resource.TfeID) (*Run, error)
	GetChunk(ctx context.Context, opts GetChunkOptions) (Chunk, error)
	TailRun(context.Context, TailOptions) (<-chan Chunk, error)
	DeleteRun(context.Context, resource.TfeID) error
	ApplyRun(context.Context, resource.TfeID) error

	GetRunPlanFile(ctx context.Context, id resource.TfeID, format PlanFormat) ([]byte, error)
	UploadRunPlanFile(ctx context.Context, id resource.TfeID, plan []byte, format PlanFormat) error

	GetLockFile(ctx context.Context, id resource.TfeID) ([]byte, error)
	UploadLockFile(ctx context.Context, id resource.TfeID, lockFile []byte) error

	PutChunk(ctx context.Context, opts PutChunkOptions) error
}

func (a *api) AddHandlers(r *mux.Router) {
	r.HandleFunc("/runs", a.list).Methods("GET")
	r.HandleFunc("/runs/{id}", a.get).Methods("GET")
	r.HandleFunc("/runs/{id}/planfile", a.getPlanFile).Methods("GET")
	r.HandleFunc("/runs/{id}/planfile", a.uploadPlanFile).Methods("PUT")
	r.HandleFunc("/runs/{id}/lockfile", a.getLockFile).Methods("GET")
	r.HandleFunc("/runs/{id}/lockfile", a.uploadLockFile).Methods("PUT")
	r.HandleFunc("/runs/{run_id}/logs/{phase}", a.putLogs).Methods("PUT")
}

func (a *api) list(w http.ResponseWriter, r *http.Request) {
	var params ListOptions
	if err := decode.All(&params, r); err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	page, err := a.Client.ListRuns(r.Context(), params)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	a.RespondWithPage(w, r, page.Items, page.Pagination)
}

func (a *api) get(w http.ResponseWriter, r *http.Request) {
	id, err := decode.ID("id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	run, err := a.Client.GetRun(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	a.Respond(w, r, run, http.StatusOK)
}

func (a *api) getPlanFile(w http.ResponseWriter, r *http.Request) {
	id, err := decode.ID("id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	opts := PlanFileOptions{}
	if err := decode.Query(&opts, r.URL.Query()); err != nil {
		tfeapi.Error(w, err)
		return
	}
	file, err := a.Client.GetRunPlanFile(r.Context(), id, opts.Format)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	if _, err := w.Write(file); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (a *api) uploadPlanFile(w http.ResponseWriter, r *http.Request) {
	id, err := decode.ID("id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	opts := PlanFileOptions{}
	if err := decode.Query(&opts, r.URL.Query()); err != nil {
		tfeapi.Error(w, err)
		return
	}
	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, r.Body); err != nil {
		tfeapi.Error(w, err)
		return
	}
	if err := a.Client.UploadRunPlanFile(r.Context(), id, buf.Bytes(), opts.Format); err != nil {
		tfeapi.Error(w, err)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func (a *api) getLockFile(w http.ResponseWriter, r *http.Request) {
	id, err := decode.ID("id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	file, err := a.Client.GetLockFile(r.Context(), id)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	if _, err := w.Write(file); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (a *api) uploadLockFile(w http.ResponseWriter, r *http.Request) {
	id, err := decode.ID("id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, r.Body); err != nil {
		tfeapi.Error(w, err)
		return
	}
	if err := a.Client.UploadLockFile(r.Context(), id, buf.Bytes()); err != nil {
		tfeapi.Error(w, err)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func (a *api) putLogs(w http.ResponseWriter, r *http.Request) {
	var opts PutChunkOptions
	if err := decode.All(&opts, r); err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, r.Body); err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	opts.Data = buf.Bytes()
	if err := a.Client.PutChunk(r.Context(), opts); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
}
