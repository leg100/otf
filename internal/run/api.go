package run

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/DataDog/jsonapi"
	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	otfapi "github.com/leg100/otf/internal/api"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/pubsub"
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
	r.HandleFunc("/runs/{run_id}", a.get).Methods("GET")
	r.HandleFunc("/runs/{id}/actions/start/{phase}", a.startPhase).Methods("POST")
	r.HandleFunc("/runs/{id}/actions/finish/{phase}", a.finishPhase).Methods("POST")
	r.HandleFunc("/runs/{id}/planfile", a.getPlanFile).Methods("GET")
	r.HandleFunc("/runs/{id}/planfile", a.uploadPlanFile).Methods("PUT")
	r.HandleFunc("/runs/{id}/lockfile", a.getLockFile).Methods("GET")
	r.HandleFunc("/runs/{id}/lockfile", a.uploadLockFile).Methods("PUT")
	r.HandleFunc("/watch", a.watch).Methods("GET")
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
	var opts PhaseFinishOptions
	if err := json.NewDecoder(r.Body).Decode(&opts); err != nil {
		tfeapi.Error(w, err)
		return
	}
	run, err := a.FinishPhase(r.Context(), params.RunID, params.Phase, opts)
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

// watch responds with a stream of run events
func (a *api) watch(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	var params WatchOptions
	if err := decode.Query(&params, r.URL.Query()); err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	events, err := a.Watch(r.Context(), params)
	if err != nil && errors.Is(err, internal.ErrAccessNotPermitted) {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "\r\n")
	flusher.Flush()

	for event := range events {
		run := event.Payload.(*Run)
		b, err := jsonapi.Marshal(run)
		if err != nil {
			a.Error(err, "marshalling run event", "event", event.Type)
			continue
		}
		pubsub.WriteSSEEvent(w, b, event.Type, true)
		flusher.Flush()
	}
}
