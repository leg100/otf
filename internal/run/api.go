package run

import (
	"bytes"
	"errors"
	"io"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/tfeapi"
)

type api struct {
	Service
	*tfeapi.Responder
}

func (a *api) addHandlers(r *mux.Router) {
	r.HandleFunc("/api/runs", a.list).Methods("GET")
	r.HandleFunc("/api/runs/{run_id}", a.get).Methods("GET")
	r.HandleFunc("/runs/{id}/actions/start/{phase}", a.startPhase).Methods("POST")
	r.HandleFunc("/runs/{id}/actions/finish/{phase}", a.finishPhase).Methods("POST")
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

func (a *api) startPhase(w http.ResponseWriter, r *http.Request) {
	var params struct {
		RunID string             `schema:"id,required"`
		Phase internal.PhaseType `schema:"phase,required"`
	}
	if err := decode.Route(&params, r); err != nil {
		tfeapi.Error(w, err)
		return
	}
	run, err := a.StartPhase(r.Context(), params.RunID, params.Phase, PhaseStartOptions{})
	if errors.Is(err, internal.ErrPhaseAlreadyStarted) {
		// A bit silly, but OTF uses the teapot status as a unique means of
		// informing the agent the phase has been started by another agent.
		w.WriteHeader(http.StatusTeapot)
		return
	} else if err != nil {
		tfeapi.Error(w, err)
		return
	}
	a.Respond(w, r, run, http.StatusOK)
}

func (a *api) finishPhase(w http.ResponseWriter, r *http.Request) {
	var params struct {
		RunID string             `schema:"id,required"`
		Phase internal.PhaseType `schema:"phase,required"`
	}
	if err := decode.Route(&params, r); err != nil {
		tfeapi.Error(w, err)
		return
	}
	run, err := a.FinishPhase(r.Context(), params.RunID, params.Phase, PhaseFinishOptions{})
	if err != nil {
		tfeapi.Error(w, err)
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
	err = a.UploadPlanFile(r.Context(), id, buf.Bytes(), opts.Format)
	if err != nil {
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
