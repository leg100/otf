package run

import (
	"bytes"
	"io"
	"net/http"

	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	otfapi "github.com/leg100/otf/internal/api"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/tfeapi"
)

type api struct {
	Service
	*tfeapi.Responder
	logr.Logger
}

func (a *api) addHandlers(r *mux.Router) {
	r = r.PathPrefix(otfapi.DefaultBasePath).Subrouter()
	r.HandleFunc("/runs", a.list).Methods("GET")
	r.HandleFunc("/runs/{id}", a.get).Methods("GET")
	r.HandleFunc("/runs/{id}/planfile", a.getPlanFile).Methods("GET")
	r.HandleFunc("/runs/{id}/planfile", a.uploadPlanFile).Methods("PUT")
	r.HandleFunc("/runs/{id}/lockfile", a.getLockFile).Methods("GET")
	r.HandleFunc("/runs/{id}/lockfile", a.uploadLockFile).Methods("PUT")
}

func (a *api) list(w http.ResponseWriter, r *http.Request) {
	var params ListOptions
	if err := decode.All(&params, r); err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	page, err := a.ListRuns(r.Context(), params)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	a.RespondWithPage(w, r, page.Items, page.Pagination)
}

func (a *api) get(w http.ResponseWriter, r *http.Request) {
	id, err := decode.Param("id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	run, err := a.GetRun(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	a.Respond(w, r, run, http.StatusOK)
}

func (a *api) getPlanFile(w http.ResponseWriter, r *http.Request) {
	id, err := decode.Param("id", r)
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
	id, err := decode.Param("id", r)
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
	id, err := decode.Param("id", r)
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
	id, err := decode.Param("id", r)
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
