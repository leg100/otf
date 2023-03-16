package run

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/leg100/otf/http/decode"
	"github.com/leg100/otf/http/jsonapi"
)

type (
	api struct {
		svc Service

		*JSONAPIMarshaler
	}

	// planFileOptions are options for the plan file API
	planFileOptions struct {
		Format PlanFormat `schema:"format,required"`
	}
)

func (h *api) addHandlers(r *mux.Router) {
	// Run routes
	r.HandleFunc("/runs", h.create).Methods("POST")
	r.HandleFunc("/runs/{id}/actions/apply", h.applyRun).Methods("POST")
	r.HandleFunc("/runs", h.list).Methods("GET")
	r.HandleFunc("/workspaces/{workspace_id}/runs", h.list).Methods("GET")
	r.HandleFunc("/runs/{id}", h.get).Methods("GET")
	r.HandleFunc("/runs/{id}/actions/discard", h.discard).Methods("POST")
	r.HandleFunc("/runs/{id}/actions/cancel", h.cancel).Methods("POST")
	r.HandleFunc("/runs/{id}/actions/force-cancel", h.forceCancel).Methods("POST")
	r.HandleFunc("/organizations/{organization_name}/runs/queue", h.getRunQueue).Methods("GET")

	// Run routes for exclusive use by remote agents
	r.HandleFunc("/runs/{id}/actions/start/{phase}", h.startPhase).Methods("POST")
	r.HandleFunc("/runs/{id}/actions/finish/{phase}", h.finishPhase).Methods("POST")
	r.HandleFunc("/runs/{run_id}/planfile", h.getPlanFile).Methods("GET")
	r.HandleFunc("/runs/{run_id}/planfile", h.uploadPlanFile).Methods("PUT")
	r.HandleFunc("/runs/{run_id}/lockfile", h.getLockFile).Methods("GET")
	r.HandleFunc("/runs/{run_id}/lockfile", h.uploadLockFile).Methods("PUT")

	// Plan routes
	r.HandleFunc("/plans/{plan_id}", h.getPlan).Methods("GET")
	r.HandleFunc("/plans/{plan_id}/json-output", h.getPlanJSON).Methods("GET")

	// Apply routes
	r.HandleFunc("/applies/{apply_id}", h.getApply).Methods("GET")
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

	if err = s.svc.cancel(r.Context(), id); err == ErrRunCancelNotAllowed {
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
