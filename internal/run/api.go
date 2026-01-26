package run

import (
	"bytes"
	"io"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	otfhttp "github.com/leg100/otf/internal/http"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/tfeapi"
)

type api struct {
	internal.Verifier // for verifying upload url
	*Service
	*tfeapi.Responder
	logr.Logger
}

func (a *api) addHandlers(r *mux.Router) {
	// client is typically terraform-cli
	signed := r.PathPrefix("/signed/{signature.expiry}").Subrouter()
	signed.Use(internal.VerifySignedURL(a.Verifier))
	signed.HandleFunc("/runs/{run_id}/logs/{phase}", a.getLogs).Methods("GET")

	r = r.PathPrefix(otfhttp.APIBasePath).Subrouter()
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
	page, err := a.List(r.Context(), params)
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
	run, err := a.Get(r.Context(), id)
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
	file, err := a.GetPlanFile(r.Context(), id, opts.Format)
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
	if err := a.UploadPlanFile(r.Context(), id, buf.Bytes(), opts.Format); err != nil {
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
	file, err := a.GetLockFile(r.Context(), id)
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
	if err := a.UploadLockFile(r.Context(), id, buf.Bytes()); err != nil {
		tfeapi.Error(w, err)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func (a *api) getLogs(w http.ResponseWriter, r *http.Request) {
	var opts GetChunkOptions
	if err := decode.All(&opts, r); err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	chunk, err := a.GetChunk(r.Context(), opts)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if _, err := w.Write(chunk.Data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
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
	if err := a.PutChunk(r.Context(), opts); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
}
