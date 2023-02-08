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

type handlers struct {
	Application appService
}

func (h *handlers) AddHandlers(r *mux.Router) {
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
	r.HandleFunc("/runs/{run_id}/logs/{phase}", h.putLogs).Methods("PUT")
	r.HandleFunc("/runs/{run_id}/planfile", h.getPlanFile).Methods("GET")
	r.HandleFunc("/runs/{run_id}/planfile", h.uploadPlanFile).Methods("PUT")
	r.HandleFunc("/runs/{run_id}/lockfile", h.getLockFile).Methods("GET")
	r.HandleFunc("/runs/{run_id}/lockfile", h.uploadLockFile).Methods("PUT")
}

func (s *handlers) CreateRun(w http.ResponseWriter, r *http.Request) {
	opts := jsonapiCreateOptions{}
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
	run, err := s.Application.CreateRun(r.Context(), opts.Workspace.ID, otf.RunCreateOptions{
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
	jsonapi.WriteResponse(w, r, &Run{run, r, s}, withCode(http.StatusCreated))
}

func (s *handlers) startPhase(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	opts := otf.PhaseStartOptions{}
	if err := jsonapi.UnmarshalPayload(r.Body, &opts); err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}
	run, err := s.Application.StartPhase(
		r.Context(),
		vars["id"],
		otf.PhaseType(vars["phase"]),
		opts)
	if err != nil {
		jsonapi.Error(w, http.StatusNotFound, err)
		return
	}
	jsonapi.WriteResponse(w, r, &Run{run, r, s})
}

func (s *handlers) finishPhase(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	opts := otf.PhaseFinishOptions{}
	if err := jsonapi.UnmarshalPayload(r.Body, &opts); err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}
	run, err := s.Application.FinishPhase(
		r.Context(),
		vars["id"],
		otf.PhaseType(vars["phase"]),
		opts)
	if err != nil {
		jsonapi.Error(w, http.StatusNotFound, err)
		return
	}
	jsonapi.WriteResponse(w, r, &Run{run, r, s})
}

func (s *handlers) GetRun(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	run, err := s.Application.GetRun(r.Context(), vars["id"])
	if err != nil {
		jsonapi.Error(w, http.StatusNotFound, err)
		return
	}
	jsonapi.WriteResponse(w, r, &Run{run, r, s})
}

func (s *handlers) ListRuns(w http.ResponseWriter, r *http.Request) {
	s.listRuns(w, r, otf.RunListOptions{})
}

func (s *handlers) GetRunsQueue(w http.ResponseWriter, r *http.Request) {
	s.listRuns(w, r, otf.RunListOptions{
		Statuses: []otf.RunStatus{otf.RunPlanQueued, otf.RunApplyQueued},
	})
}

func (s *handlers) listRuns(w http.ResponseWriter, r *http.Request, opts otf.RunListOptions) {
	if err := decode.Query(&opts, r.URL.Query()); err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}
	if err := decode.Route(&opts, r); err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}
	rl, err := s.Application.ListRuns(r.Context(), opts)
	if err != nil {
		jsonapi.Error(w, http.StatusNotFound, err)
		return
	}
	jsonapi.WriteResponse(w, r, &RunList{rl, r, s})
}

func (s *handlers) ApplyRun(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	opts := otf.RunApplyOptions{}
	if err := jsonapi.UnmarshalPayload(r.Body, &opts); err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}
	if err := s.Application.ApplyRun(r.Context(), vars["id"], opts); err != nil {
		jsonapi.Error(w, http.StatusNotFound, err)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func (s *handlers) DiscardRun(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	opts := otf.RunDiscardOptions{}
	if err := jsonapi.UnmarshalPayload(r.Body, &opts); err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}
	err := s.Application.DiscardRun(r.Context(), vars["id"], opts)
	if err == otf.ErrRunDiscardNotAllowed {
		jsonapi.Error(w, http.StatusConflict, err)
		return
	} else if err != nil {
		jsonapi.Error(w, http.StatusNotFound, err)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func (s *handlers) CancelRun(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	opts := otf.RunCancelOptions{}
	if err := jsonapi.UnmarshalPayload(r.Body, &opts); err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}
	err := s.Application.CancelRun(r.Context(), vars["id"], opts)
	if err == otf.ErrRunCancelNotAllowed {
		jsonapi.Error(w, http.StatusConflict, err)
		return
	} else if err != nil {
		jsonapi.Error(w, http.StatusNotFound, err)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func (s *handlers) ForceCancelRun(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	opts := otf.RunForceCancelOptions{}
	if err := jsonapi.UnmarshalPayload(r.Body, &opts); err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}
	err := s.Application.ForceCancelRun(r.Context(), vars["id"], opts)
	if err == otf.ErrRunForceCancelNotAllowed {
		jsonapi.Error(w, http.StatusConflict, err)
		return
	} else if err != nil {
		jsonapi.Error(w, http.StatusNotFound, err)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func (s *handlers) getPlanFile(w http.ResponseWriter, r *http.Request) {
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

func (s *handlers) uploadPlanFile(w http.ResponseWriter, r *http.Request) {
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

func (s *handlers) getLockFile(w http.ResponseWriter, r *http.Request) {
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

func (s *handlers) uploadLockFile(w http.ResponseWriter, r *http.Request) {
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
func (s *handlers) getPlan(w http.ResponseWriter, r *http.Request) {
	// otf's plan IDs are simply the corresponding run ID
	planID := mux.Vars(r)["plan_id"]
	runID := otf.ConvertID(planID, "run")

	run, err := s.Application.GetRun(r.Context(), runID)
	if err != nil {
		jsonapi.Error(w, http.StatusNotFound, err)
		return
	}
	jsonapi.WriteResponse(w, r, &plan{run.Plan(), r, s})
}

// getPlanJSON retrieves a plan object's plan file in JSON format.
//
// https://www.terraform.io/cloud-docs/api-docs/plans#retrieve-the-json-execution-plan
func (s *handlers) getPlanJSON(w http.ResponseWriter, r *http.Request) {
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

func (s *handlers) GetApply(w http.ResponseWriter, r *http.Request) {
	applyID := mux.Vars(r)["apply_id"]
	runID := otf.ConvertID(applyID, "run")

	run, err := s.Application.GetRun(r.Context(), runID)
	if err != nil {
		jsonapi.Error(w, http.StatusNotFound, err)
		return
	}
	jsonapi.WriteResponse(w, r, &apply{run.Apply(), r, s})
}

type apply struct {
	*otf.Apply
	req *http.Request
	*handlers
}

type RunList struct {
	*otf.RunList
	req *http.Request
	*handlers
}

// ToJSONAPI assembles a JSON-API DTO.
func (l *RunList) ToJSONAPI() any {
	obj := &jsonapi.RunList{
		Pagination: l.Pagination.ToJSONAPI(),
	}
	for _, item := range l.Items {
		obj.Items = append(obj.Items, (&Run{item, l.req, l.Server}).ToJSONAPI().(*jsonapi.Run))
	}
	return obj
}
