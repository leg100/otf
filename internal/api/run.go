package api

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/DataDog/jsonapi"
	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/api/types"
	otfhttp "github.com/leg100/otf/internal/http"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/pubsub"
	"github.com/leg100/otf/internal/run"
)

func (a *api) addRunHandlers(r *mux.Router) {
	r = otfhttp.APIRouter(r)

	// Run routes
	r.HandleFunc("/runs", a.createRun).Methods("POST")
	r.HandleFunc("/runs/{id}/actions/apply", a.applyRun).Methods("POST")
	r.HandleFunc("/runs", a.listRuns).Methods("GET")
	r.HandleFunc("/workspaces/{workspace_id}/runs", a.listRuns).Methods("GET")
	r.HandleFunc("/runs/{id}", a.getRun).Methods("GET")
	r.HandleFunc("/runs/{id}/actions/discard", a.discardRun).Methods("POST")
	r.HandleFunc("/runs/{id}/actions/cancel", a.cancelRun).Methods("POST")
	r.HandleFunc("/runs/{id}/actions/force-cancel", a.forceCancelRun).Methods("POST")
	r.HandleFunc("/organizations/{organization_name}/runs/queue", a.getRunQueue).Methods("GET")
	r.HandleFunc("/watch", a.watchRun).Methods("GET")

	// Run routes for exclusive use by remote agents
	r.HandleFunc("/runs/{id}/actions/start/{phase}", a.startPhase).Methods("POST")
	r.HandleFunc("/runs/{id}/actions/finish/{phase}", a.finishPhase).Methods("POST")
	r.HandleFunc("/runs/{id}/planfile", a.getPlanFile).Methods("GET")
	r.HandleFunc("/runs/{id}/planfile", a.uploadPlanFile).Methods("PUT")
	r.HandleFunc("/runs/{id}/lockfile", a.getLockFile).Methods("GET")
	r.HandleFunc("/runs/{id}/lockfile", a.uploadLockFile).Methods("PUT")

	// Plan routes
	r.HandleFunc("/plans/{plan_id}", a.getPlan).Methods("GET")
	r.HandleFunc("/plans/{plan_id}/json-output", a.getPlanJSON).Methods("GET")

	// Apply routes
	r.HandleFunc("/applies/{apply_id}", a.getApply).Methods("GET")
}

func (a *api) createRun(w http.ResponseWriter, r *http.Request) {
	var opts types.RunCreateOptions
	if err := unmarshal(r.Body, &opts); err != nil {
		Error(w, err)
		return
	}
	if opts.Workspace == nil {
		Error(w, &internal.MissingParameterError{Parameter: "workspace"})
		return
	}
	var configurationVersionID *string
	if opts.ConfigurationVersion != nil {
		configurationVersionID = &opts.ConfigurationVersion.ID
	}
	run, err := a.CreateRun(r.Context(), opts.Workspace.ID, run.RunCreateOptions{
		AutoApply:              opts.AutoApply,
		IsDestroy:              opts.IsDestroy,
		Refresh:                opts.Refresh,
		RefreshOnly:            opts.RefreshOnly,
		Message:                opts.Message,
		ConfigurationVersionID: configurationVersionID,
		TargetAddrs:            opts.TargetAddrs,
		ReplaceAddrs:           opts.ReplaceAddrs,
		PlanOnly:               opts.PlanOnly,
	})
	if err != nil {
		Error(w, err)
		return
	}

	a.writeResponse(w, r, run, withCode(http.StatusCreated))
}

func (a *api) startPhase(w http.ResponseWriter, r *http.Request) {
	var params struct {
		RunID string             `schema:"id,required"`
		Phase internal.PhaseType `schema:"phase,required"`
	}
	if err := decode.Route(&params, r); err != nil {
		Error(w, err)
		return
	}

	run, err := a.StartPhase(r.Context(), params.RunID, params.Phase, run.PhaseStartOptions{})
	if err != nil {
		Error(w, err)
		return
	}

	a.writeResponse(w, r, run)
}

func (a *api) finishPhase(w http.ResponseWriter, r *http.Request) {
	var params struct {
		RunID string             `schema:"id,required"`
		Phase internal.PhaseType `schema:"phase,required"`
	}
	if err := decode.Route(&params, r); err != nil {
		Error(w, err)
		return
	}

	run, err := a.FinishPhase(r.Context(), params.RunID, params.Phase, run.PhaseFinishOptions{})
	if err != nil {
		Error(w, err)
		return
	}

	a.writeResponse(w, r, run)
}

func (a *api) getRun(w http.ResponseWriter, r *http.Request) {
	id, err := decode.Param("id", r)
	if err != nil {
		Error(w, err)
		return
	}

	run, err := a.GetRun(r.Context(), id)
	if err != nil {
		Error(w, err)
		return
	}

	a.writeResponse(w, r, run)
}

func (a *api) listRuns(w http.ResponseWriter, r *http.Request) {
	a.listRunsWithOptions(w, r, run.RunListOptions{})
}

func (a *api) getRunQueue(w http.ResponseWriter, r *http.Request) {
	a.listRunsWithOptions(w, r, run.RunListOptions{
		Statuses: []internal.RunStatus{internal.RunPlanQueued, internal.RunApplyQueued},
	})
}

func (a *api) listRunsWithOptions(w http.ResponseWriter, r *http.Request, opts run.RunListOptions) {
	if err := decode.All(&opts, r); err != nil {
		Error(w, err)
		return
	}

	list, err := a.ListRuns(r.Context(), opts)
	if err != nil {
		Error(w, err)
		return
	}

	a.writeResponse(w, r, list)
}

func (a *api) applyRun(w http.ResponseWriter, r *http.Request) {
	id, err := decode.Param("id", r)
	if err != nil {
		Error(w, err)
		return
	}

	if err := a.Apply(r.Context(), id); err != nil {
		Error(w, err)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func (a *api) discardRun(w http.ResponseWriter, r *http.Request) {
	id, err := decode.Param("id", r)
	if err != nil {
		Error(w, err)
		return
	}

	if err = a.DiscardRun(r.Context(), id); err != nil {
		Error(w, err)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func (a *api) cancelRun(w http.ResponseWriter, r *http.Request) {
	id, err := decode.Param("id", r)
	if err != nil {
		Error(w, err)
		return
	}

	if _, err = a.Cancel(r.Context(), id); err != nil {
		Error(w, err)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func (a *api) forceCancelRun(w http.ResponseWriter, r *http.Request) {
	id, err := decode.Param("id", r)
	if err != nil {
		Error(w, err)
		return
	}

	if err := a.ForceCancelRun(r.Context(), id); err != nil {
		Error(w, err)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func (a *api) getPlanFile(w http.ResponseWriter, r *http.Request) {
	id, err := decode.Param("id", r)
	if err != nil {
		Error(w, err)
		return
	}
	opts := run.PlanFileOptions{}
	if err := decode.Query(&opts, r.URL.Query()); err != nil {
		Error(w, err)
		return
	}

	file, err := a.GetPlanFile(r.Context(), id, opts.Format)
	if err != nil {
		Error(w, err)
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
		Error(w, err)
		return
	}
	opts := run.PlanFileOptions{}
	if err := decode.Query(&opts, r.URL.Query()); err != nil {
		Error(w, err)
		return
	}

	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, r.Body); err != nil {
		Error(w, err)
		return
	}

	err = a.UploadPlanFile(r.Context(), id, buf.Bytes(), opts.Format)
	if err != nil {
		Error(w, err)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func (a *api) getLockFile(w http.ResponseWriter, r *http.Request) {
	id, err := decode.Param("id", r)
	if err != nil {
		Error(w, err)
		return
	}

	file, err := a.GetLockFile(r.Context(), id)
	if err != nil {
		Error(w, err)
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
		Error(w, err)
		return
	}

	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, r.Body); err != nil {
		Error(w, err)
		return
	}

	err = a.UploadLockFile(r.Context(), id, buf.Bytes())
	if err != nil {
		Error(w, err)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

// These endpoints implement the documented plan API:
//
// https://www.terraform.io/cloud-docs/api-docs/plans#retrieve-the-json-execution-plan
//

// getPlan retrieves a plan object in JSON-API format.
//
// https://www.terraform.io/cloud-docs/api-docs/plans#show-a-plan
func (a *api) getPlan(w http.ResponseWriter, r *http.Request) {
	id, err := decode.Param("plan_id", r)
	if err != nil {
		Error(w, err)
		return
	}

	// otf's plan IDs are simply the corresponding run ID
	run, err := a.GetRun(r.Context(), internal.ConvertID(id, "run"))
	if err != nil {
		Error(w, err)
		return
	}

	a.writeResponse(w, r, run.Plan)
}

// getPlanJSON retrieves a plan object's plan file in JSON format.
//
// https://www.terraform.io/cloud-docs/api-docs/plans#retrieve-the-json-execution-plan
func (a *api) getPlanJSON(w http.ResponseWriter, r *http.Request) {
	id, err := decode.Param("plan_id", r)
	if err != nil {
		Error(w, err)
		return
	}

	// otf's plan IDs are simply the corresponding run ID
	json, err := a.GetPlanFile(r.Context(), internal.ConvertID(id, "run"), run.PlanFormatJSON)
	if err != nil {
		Error(w, err)
		return
	}
	if _, err := w.Write(json); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (a *api) getApply(w http.ResponseWriter, r *http.Request) {
	id, err := decode.Param("apply_id", r)
	if err != nil {
		Error(w, err)
		return
	}

	// otf's apply IDs are simply the corresponding run ID
	run, err := a.GetRun(r.Context(), internal.ConvertID(id, "run"))
	if err != nil {
		Error(w, err)
		return
	}

	a.writeResponse(w, r, run.Apply)
}

// Watch handler responds with a stream of run events
func (a *api) watchRun(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	var params run.WatchOptions
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
		run := event.Payload.(*run.Run)
		jrun, _, err := a.toRun(run, r)
		if err != nil {
			a.Error(err, "marshalling run event", "event", event.Type)
			continue
		}
		b, err := jsonapi.Marshal(jrun)
		if err != nil {
			a.Error(err, "marshalling run event", "event", event.Type)
			continue
		}
		pubsub.WriteSSEEvent(w, b, event.Type, true)
		flusher.Flush()
	}
}
