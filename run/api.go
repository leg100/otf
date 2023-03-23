package run

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	otfhttp "github.com/leg100/otf/http"
	"github.com/leg100/otf/http/decode"
	"github.com/leg100/otf/http/jsonapi"
)

type (
	api struct {
		logr.Logger
		jsonapiMarshaler

		svc Service
	}

	jsonapiMarshaler interface {
		MarshalJSONAPI(run *Run, r *http.Request) ([]byte, error)
		toRun(run *Run, r *http.Request) (*jsonapi.Run, error)
		toList(list *RunList, r *http.Request) (*jsonapi.RunList, error)
		toPhase(from Phase, r *http.Request) (any, error)
	}

	// planFileOptions are options for the plan file API
	planFileOptions struct {
		Format PlanFormat `schema:"format,required"`
	}
)

func (s *api) addHandlers(r *mux.Router) {
	r = otfhttp.APIRouter(r)

	// Run routes
	r.HandleFunc("/runs", s.create).Methods("POST")
	r.HandleFunc("/runs/{id}/actions/apply", s.applyRun).Methods("POST")
	r.HandleFunc("/runs", s.list).Methods("GET")
	r.HandleFunc("/workspaces/{workspace_id}/runs", s.list).Methods("GET")
	r.HandleFunc("/runs/{id}", s.get).Methods("GET")
	r.HandleFunc("/runs/{id}/actions/discard", s.discard).Methods("POST")
	r.HandleFunc("/runs/{id}/actions/cancel", s.cancel).Methods("POST")
	r.HandleFunc("/runs/{id}/actions/force-cancel", s.forceCancel).Methods("POST")
	r.HandleFunc("/organizations/{organization_name}/runs/queue", s.getRunQueue).Methods("GET")
	r.HandleFunc("/watch", s.watch).Methods("GET")

	// Run routes for exclusive use by remote agents
	r.HandleFunc("/runs/{id}/actions/start/{phase}", s.startPhase).Methods("POST")
	r.HandleFunc("/runs/{id}/actions/finish/{phase}", s.finishPhase).Methods("POST")
	r.HandleFunc("/runs/{run_id}/planfile", s.getPlanFile).Methods("GET")
	r.HandleFunc("/runs/{run_id}/planfile", s.uploadPlanFile).Methods("PUT")
	r.HandleFunc("/runs/{run_id}/lockfile", s.getLockFile).Methods("GET")
	r.HandleFunc("/runs/{run_id}/lockfile", s.uploadLockFile).Methods("PUT")

	// Plan routes
	r.HandleFunc("/plans/{plan_id}", s.getPlan).Methods("GET")
	r.HandleFunc("/plans/{plan_id}/json-output", s.getPlanJSON).Methods("GET")

	// Apply routes
	r.HandleFunc("/applies/{apply_id}", s.getApply).Methods("GET")
}

func (s *api) create(w http.ResponseWriter, r *http.Request) {
	opts := jsonapi.RunCreateOptions{}
	if err := jsonapi.UnmarshalPayload(r.Body, &opts); err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}
	if opts.Workspace == nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, fmt.Errorf("missing workspace"))
		return
	}
	var configurationVersionID *string
	if opts.ConfigurationVersion != nil {
		configurationVersionID = &opts.ConfigurationVersion.ID
	}
	run, err := s.svc.CreateRun(r.Context(), opts.Workspace.ID, RunCreateOptions{
		AutoApply:              opts.AutoApply,
		IsDestroy:              opts.IsDestroy,
		Refresh:                opts.Refresh,
		RefreshOnly:            opts.RefreshOnly,
		Message:                opts.Message,
		ConfigurationVersionID: configurationVersionID,
		TargetAddrs:            opts.TargetAddrs,
		ReplaceAddrs:           opts.ReplaceAddrs,
	})
	if err != nil {
		jsonapi.Error(w, http.StatusNotFound, err)
		return
	}

	s.writeResponse(w, r, run, jsonapi.WithCode(http.StatusCreated))
}

func (s *api) startPhase(w http.ResponseWriter, r *http.Request) {
	params := struct {
		RunID string        `schema:"id,required"`
		Phase otf.PhaseType `schema:"phase,required"`
	}{}
	if err := decode.Route(&params, r); err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}

	run, err := s.svc.StartPhase(r.Context(), params.RunID, params.Phase, PhaseStartOptions{})
	if err != nil {
		jsonapi.Error(w, http.StatusNotFound, err)
		return
	}

	s.writeResponse(w, r, run)
}

func (s *api) finishPhase(w http.ResponseWriter, r *http.Request) {
	params := struct {
		RunID string        `schema:"id,required"`
		Phase otf.PhaseType `schema:"phase,required"`
	}{}
	if err := decode.Route(&params, r); err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}

	run, err := s.svc.FinishPhase(r.Context(), params.RunID, params.Phase, PhaseFinishOptions{})
	if err != nil {
		jsonapi.Error(w, http.StatusNotFound, err)
		return
	}

	s.writeResponse(w, r, run)
}

func (s *api) get(w http.ResponseWriter, r *http.Request) {
	id, err := decode.Param("id", r)
	if err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}

	run, err := s.svc.get(r.Context(), id)
	if err != nil {
		jsonapi.Error(w, http.StatusNotFound, err)
		return
	}

	s.writeResponse(w, r, run)
}

func (s *api) list(w http.ResponseWriter, r *http.Request) {
	s.listRuns(w, r, RunListOptions{})
}

func (s *api) getRunQueue(w http.ResponseWriter, r *http.Request) {
	s.listRuns(w, r, RunListOptions{
		Statuses: []otf.RunStatus{otf.RunPlanQueued, otf.RunApplyQueued},
	})
}

func (s *api) listRuns(w http.ResponseWriter, r *http.Request, opts RunListOptions) {
	if err := decode.All(&opts, r); err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}

	list, err := s.svc.ListRuns(r.Context(), opts)
	if err != nil {
		jsonapi.Error(w, http.StatusNotFound, err)
		return
	}

	s.writeResponse(w, r, list)
}

func (s *api) applyRun(w http.ResponseWriter, r *http.Request) {
	id, err := decode.Param("id", r)
	if err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}

	if err := s.svc.apply(r.Context(), id); err != nil {
		jsonapi.Error(w, http.StatusNotFound, err)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func (s *api) discard(w http.ResponseWriter, r *http.Request) {
	id, err := decode.Param("id", r)
	if err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}

	if err = s.svc.discard(r.Context(), id); err == ErrRunDiscardNotAllowed {
		jsonapi.Error(w, http.StatusConflict, err)
		return
	} else if err != nil {
		jsonapi.Error(w, http.StatusNotFound, err)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func (s *api) cancel(w http.ResponseWriter, r *http.Request) {
	id, err := decode.Param("id", r)
	if err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}

	_, err = s.svc.Cancel(r.Context(), id)
	if err == ErrRunCancelNotAllowed {
		jsonapi.Error(w, http.StatusConflict, err)
		return
	} else if err != nil {
		jsonapi.Error(w, http.StatusNotFound, err)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func (s *api) forceCancel(w http.ResponseWriter, r *http.Request) {
	id, err := decode.Param("id", r)
	if err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}

	err = s.svc.forceCancel(r.Context(), id)
	if err == ErrRunForceCancelNotAllowed {
		jsonapi.Error(w, http.StatusConflict, err)
		return
	} else if err != nil {
		jsonapi.Error(w, http.StatusNotFound, err)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func (s *api) getPlanFile(w http.ResponseWriter, r *http.Request) {
	id, err := decode.Param("id", r)
	if err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}
	opts := planFileOptions{}
	if err := decode.Query(&opts, r.URL.Query()); err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}

	file, err := s.svc.GetPlanFile(r.Context(), id, opts.Format)
	if err != nil {
		jsonapi.Error(w, http.StatusNotFound, err)
		return
	}

	if _, err := w.Write(file); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *api) uploadPlanFile(w http.ResponseWriter, r *http.Request) {
	id, err := decode.Param("id", r)
	if err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}
	opts := planFileOptions{}
	if err := decode.Query(&opts, r.URL.Query()); err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}

	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, r.Body); err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}

	err = s.svc.UploadPlanFile(r.Context(), id, buf.Bytes(), opts.Format)
	if err != nil {
		jsonapi.Error(w, http.StatusNotFound, err)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func (s *api) getLockFile(w http.ResponseWriter, r *http.Request) {
	id, err := decode.Param("id", r)
	if err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}

	file, err := s.svc.GetLockFile(r.Context(), id)
	if err != nil {
		jsonapi.Error(w, http.StatusNotFound, err)
		return
	}

	if _, err := w.Write(file); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *api) uploadLockFile(w http.ResponseWriter, r *http.Request) {
	id, err := decode.Param("id", r)
	if err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}

	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, r.Body); err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}

	err = s.svc.UploadLockFile(r.Context(), id, buf.Bytes())
	if err != nil {
		jsonapi.Error(w, http.StatusNotFound, err)
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
func (s *api) getPlan(w http.ResponseWriter, r *http.Request) {
	id, err := decode.Param("plan_id", r)
	if err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}

	// otf's plan IDs are simply the corresponding run ID
	run, err := s.svc.get(r.Context(), otf.ConvertID(id, "run"))
	if err != nil {
		jsonapi.Error(w, http.StatusNotFound, err)
		return
	}

	s.writeResponse(w, r, run.Plan)
}

// getPlanJSON retrieves a plan object's plan file in JSON format.
//
// https://www.terraform.io/cloud-docs/api-docs/plans#retrieve-the-json-execution-plan
func (s *api) getPlanJSON(w http.ResponseWriter, r *http.Request) {
	id, err := decode.Param("plan_id", r)
	if err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}

	// otf's plan IDs are simply the corresponding run ID
	json, err := s.svc.GetPlanFile(r.Context(), otf.ConvertID(id, "run"), PlanFormatJSON)
	if err != nil {
		jsonapi.Error(w, http.StatusNotFound, err)
		return
	}
	if _, err := w.Write(json); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *api) getApply(w http.ResponseWriter, r *http.Request) {
	id, err := decode.Param("apply_id", r)
	if err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}

	// otf's apply IDs are simply the corresponding run ID
	run, err := s.svc.get(r.Context(), otf.ConvertID(id, "run"))
	if err != nil {
		jsonapi.Error(w, http.StatusNotFound, err)
		return
	}

	s.writeResponse(w, r, run.Apply)
}

// Watch handler responds with a stream of events, using json encoding.
//
// NOTE: Only run events are currently supported.
func (s *api) watch(w http.ResponseWriter, r *http.Request) {
	// TODO: populate watch options
	events, err := s.svc.Watch(r.Context(), WatchOptions{})
	if err != nil {
		jsonapi.Error(w, http.StatusInternalServerError, err)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)
	rc := http.NewResponseController(w)
	rc.Flush()

	for {
		select {
		case event, ok := <-events:
			if !ok {
				// server closed connection
				return
			}

			run := event.Payload.(*Run)
			data, err := s.MarshalJSONAPI(run, r)
			if err != nil {
				s.Error(err, "marshalling run event", "event", event.Type)
				continue
			}
			otf.WriteSSEEvent(w, data, event.Type)
			rc.Flush()
		case <-r.Context().Done():
			// client closed connection
			return
		}
	}
}

// writeResponse encodes v as json:api and writes it to the body of the http response.
func (s *api) writeResponse(w http.ResponseWriter, r *http.Request, v any, opts ...func(http.ResponseWriter)) {
	var payload any
	var err error

	switch v := v.(type) {
	case *RunList:
		payload, err = s.toList(v, r)
	case *Run:
		payload, err = s.toRun(v, r)
	case Phase:
		payload, err = s.toPhase(v, r)
	}
	if err != nil {
		jsonapi.Error(w, http.StatusInternalServerError, err)
		return
	}
	jsonapi.WriteResponse(w, r, payload, opts...)
}
