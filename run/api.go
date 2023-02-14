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

type api struct {
	app app

	jsonapiMarshaler
}

func (h *api) addHandlers(r *mux.Router) {
	// Run routes
	r.HandleFunc("/runs", h.CreateRun).Methods("POST")
	r.HandleFunc("/runs/{id}/actions/apply", h.ApplyRun).Methods("POST")
	r.HandleFunc("/runs", h.ListRuns).Methods("GET")
	r.HandleFunc("/workspaces/{workspace_id}/runs", h.ListRuns).Methods("GET")
	r.HandleFunc("/runs/{id}", h.GetRun).Methods("GET")
	r.HandleFunc("/runs/{id}/actions/discard", h.DiscardRun).Methods("POST")
	r.HandleFunc("/runs/{id}/actions/cancel", h.CancelRun).Methods("POST")
	r.HandleFunc("/runs/{id}/actions/force-cancel", h.ForceCancelRun).Methods("POST")
	r.HandleFunc("/organizations/{organization_name}/runs/queue", h.GetRunsQueue).Methods("GET")

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
	r.HandleFunc("/applies/{apply_id}", h.GetApply).Methods("GET")
}

func (s *api) CreateRun(w http.ResponseWriter, r *http.Request) {
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
	run, err := s.app.CreateRun(r.Context(), opts.Workspace.ID, RunCreateOptions{
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

	jrun, err := s.toJSONAPI(run, r)
	if err != nil {
		jsonapi.Error(w, http.StatusNotFound, err)
		return
	}
	jsonapi.WriteResponse(w, r, jrun, jsonapi.WithCode(http.StatusCreated))
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

	run, err := s.app.StartPhase(r.Context(), params.RunID, params.Phase, otf.PhaseStartOptions{})
	if err != nil {
		jsonapi.Error(w, http.StatusNotFound, err)
		return
	}

	jrun, err := s.toJSONAPI(run, r)
	if err != nil {
		jsonapi.Error(w, http.StatusNotFound, err)
		return
	}
	jsonapi.WriteResponse(w, r, jrun)
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

	run, err := s.app.FinishPhase(r.Context(), params.RunID, params.Phase, otf.PhaseFinishOptions{})
	if err != nil {
		jsonapi.Error(w, http.StatusNotFound, err)
		return
	}

	jrun, err := s.toJSONAPI(run, r)
	if err != nil {
		jsonapi.Error(w, http.StatusNotFound, err)
		return
	}
	jsonapi.WriteResponse(w, r, jrun)
}

func (s *api) GetRun(w http.ResponseWriter, r *http.Request) {
	id, err := decode.Param("id", r)
	if err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}

	run, err := s.app.GetRun(r.Context(), id)
	if err != nil {
		jsonapi.Error(w, http.StatusNotFound, err)
		return
	}

	jrun, err := s.toJSONAPI(run, r)
	if err != nil {
		jsonapi.Error(w, http.StatusNotFound, err)
		return
	}
	jsonapi.WriteResponse(w, r, jrun)
}

func (s *api) ListRuns(w http.ResponseWriter, r *http.Request) {
	s.listRuns(w, r, otf.RunListOptions{})
}

func (s *api) GetRunsQueue(w http.ResponseWriter, r *http.Request) {
	s.listRuns(w, r, otf.RunListOptions{
		Statuses: []otf.RunStatus{otf.RunPlanQueued, otf.RunApplyQueued},
	})
}

func (s *api) listRuns(w http.ResponseWriter, r *http.Request, opts otf.RunListOptions) {
	if err := decode.All(&opts, r); err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}

	list, err := s.app.ListRuns(r.Context(), opts)
	if err != nil {
		jsonapi.Error(w, http.StatusNotFound, err)
		return
	}

	jsonapi.WriteResponse(w, r, &RunList{list, r, s})
}

func (s *api) ApplyRun(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	opts := otf.RunApplyOptions{}
	if err := jsonapi.UnmarshalPayload(r.Body, &opts); err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}
	if err := s.app.ApplyRun(r.Context(), vars["id"], opts); err != nil {
		jsonapi.Error(w, http.StatusNotFound, err)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func (s *api) DiscardRun(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	opts := otf.RunDiscardOptions{}
	if err := jsonapi.UnmarshalPayload(r.Body, &opts); err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}
	err := s.app.DiscardRun(r.Context(), vars["id"], opts)
	if err == otf.ErrRunDiscardNotAllowed {
		jsonapi.Error(w, http.StatusConflict, err)
		return
	} else if err != nil {
		jsonapi.Error(w, http.StatusNotFound, err)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func (s *api) CancelRun(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	opts := otf.RunCancelOptions{}
	if err := jsonapi.UnmarshalPayload(r.Body, &opts); err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}
	err := s.app.CancelRun(r.Context(), vars["id"], opts)
	if err == otf.ErrRunCancelNotAllowed {
		jsonapi.Error(w, http.StatusConflict, err)
		return
	} else if err != nil {
		jsonapi.Error(w, http.StatusNotFound, err)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func (s *api) ForceCancelRun(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	opts := otf.RunForceCancelOptions{}
	if err := jsonapi.UnmarshalPayload(r.Body, &opts); err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}
	err := s.app.ForceCancelRun(r.Context(), vars["id"], opts)
	if err == otf.ErrRunForceCancelNotAllowed {
		jsonapi.Error(w, http.StatusConflict, err)
		return
	} else if err != nil {
		jsonapi.Error(w, http.StatusNotFound, err)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func (s *api) getPlanFile(w http.ResponseWriter, r *http.Request) {
	opts := planFileOptions{}
	if err := decode.Query(&opts, r.URL.Query()); err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}
	vars := mux.Vars(r)
	file, err := s.GetPlanFile(r.Context(), vars["run_id"], opts.Format)
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
	opts := planFileOptions{}
	if err := decode.Query(&opts, r.URL.Query()); err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}
	vars := mux.Vars(r)
	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, r.Body); err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}
	err := s.UploadPlanFile(r.Context(), vars["run_id"], buf.Bytes(), opts.Format)
	if err != nil {
		jsonapi.Error(w, http.StatusNotFound, err)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func (s *api) getLockFile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	file, err := s.GetLockFile(r.Context(), vars["run_id"])
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
	vars := mux.Vars(r)
	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, r.Body); err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}
	err := s.UploadLockFile(r.Context(), vars["run_id"], buf.Bytes())
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
//
func (s *api) getPlan(w http.ResponseWriter, r *http.Request) {
	// otf's plan IDs are simply the corresponding run ID
	planID := mux.Vars(r)["plan_id"]
	runID := otf.ConvertID(planID, "run")

	run, err := s.app.GetRun(r.Context(), runID)
	if err != nil {
		jsonapi.Error(w, http.StatusNotFound, err)
		return
	}
	jsonapi.WriteResponse(w, r, &plan{run.Plan(), r, s})
}

// getPlanJSON retrieves a plan object's plan file in JSON format.
//
// https://www.terraform.io/cloud-docs/api-docs/plans#retrieve-the-json-execution-plan
func (s *api) getPlanJSON(w http.ResponseWriter, r *http.Request) {
	// otf's plan IDs are simply the corresponding run ID
	planID := mux.Vars(r)["plan_id"]
	runID := otf.ConvertID(planID, "run")

	json, err := s.GetPlanFile(r.Context(), runID, otf.PlanFormatJSON)
	if err != nil {
		jsonapi.Error(w, http.StatusNotFound, err)
		return
	}
	if _, err := w.Write(json); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *api) GetApply(w http.ResponseWriter, r *http.Request) {
	applyID := mux.Vars(r)["apply_id"]
	runID := otf.ConvertID(applyID, "run")

	run, err := s.app.GetRun(r.Context(), runID)
	if err != nil {
		jsonapi.Error(w, http.StatusNotFound, err)
		return
	}
	jsonapi.WriteResponse(w, r, &apply{run.Apply(), r, s})
}
