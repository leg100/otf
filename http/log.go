package http

import (
	"bytes"
	"io"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/jsonapi"
	"github.com/leg100/otf"
)

func (s *Server) GetPlanLogs(w http.ResponseWriter, r *http.Request) {
	opts := otf.RunCreateOptions{}
	if err := jsonapi.UnmarshalPayload(r.Body, &opts); err != nil {
		WriteError(w, http.StatusUnprocessableEntity, err)
		return
	}

	obj, err := s.RunService.Create(r.Context(), opts)
	if err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	WriteResponse(w, r, RunJSONAPIObject(r, obj), WithCode(http.StatusCreated))
}

func (s *Server) GetRun(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	obj, err := s.RunService.Get(vars["id"])
	if err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	WriteResponse(w, r, RunJSONAPIObject(r, obj))
}

func (s *Server) ListRuns(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	var opts otf.RunListOptions
	if err := DecodeQuery(&opts, r.URL.Query()); err != nil {
		WriteError(w, http.StatusUnprocessableEntity, err)
		return
	}

	workspaceID := vars["workspace_id"]
	opts.WorkspaceID = &workspaceID

	obj, err := s.RunService.List(opts)
	if err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	WriteResponse(w, r, RunListJSONAPIObject(r, obj))
}

func (s *Server) UploadLogs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, r.Body); err != nil {
		WriteError(w, http.StatusUnprocessableEntity, err)
		return
	}

	var opts otf.RunUploadLogsOptions

	if err := DecodeQuery(&opts, r.URL.Query()); err != nil {
		WriteError(w, http.StatusUnprocessableEntity, err)
		return
	}

	if err := s.RunService.UploadLogs(r.Context(), vars["id"], buf.Bytes(), opts); err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}
}

func (s *Server) UploadPlanFile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, r.Body); err != nil {
		WriteError(w, http.StatusUnprocessableEntity, err)
		return
	}

	var opts otf.PlanFileOptions

	if err := DecodeQuery(&opts, r.URL.Query()); err != nil {
		WriteError(w, http.StatusUnprocessableEntity, err)
		return
	}

	if err := s.RunService.UploadPlanFile(r.Context(), vars["id"], buf.Bytes(), opts); err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}
}

func (s *Server) ApplyRun(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	opts := otf.RunApplyOptions{}
	if err := jsonapi.UnmarshalPayload(r.Body, &opts); err != nil {
		WriteError(w, http.StatusUnprocessableEntity, err)
		return
	}

	if err := s.RunService.Apply(vars["id"], opts); err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func (s *Server) DiscardRun(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	opts := otf.RunDiscardOptions{}
	if err := jsonapi.UnmarshalPayload(r.Body, &opts); err != nil {
		WriteError(w, http.StatusUnprocessableEntity, err)
		return
	}

	err := s.RunService.Discard(vars["id"], opts)
	if err == otf.ErrRunDiscardNotAllowed {
		WriteError(w, http.StatusConflict, err)
		return
	} else if err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func (s *Server) CancelRun(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	opts := otf.RunCancelOptions{}
	if err := jsonapi.UnmarshalPayload(r.Body, &opts); err != nil {
		WriteError(w, http.StatusUnprocessableEntity, err)
		return
	}

	err := s.RunService.Cancel(vars["id"], opts)
	if err == otf.ErrRunCancelNotAllowed {
		WriteError(w, http.StatusConflict, err)
		return
	} else if err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func (s *Server) ForceCancelRun(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	opts := otf.RunForceCancelOptions{}
	if err := jsonapi.UnmarshalPayload(r.Body, &opts); err != nil {
		WriteError(w, http.StatusUnprocessableEntity, err)
		return
	}

	err := s.RunService.ForceCancel(vars["id"], opts)
	if err == otf.ErrRunForceCancelNotAllowed {
		WriteError(w, http.StatusConflict, err)
		return
	} else if err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func (s *Server) GetPlanFile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	var opts otf.PlanFileOptions

	if err := DecodeQuery(&opts, r.URL.Query()); err != nil {
		WriteError(w, http.StatusUnprocessableEntity, err)
		return
	}

	s.getPlanFile(w, r, vars["id"], opts)
}

func (s *Server) GetJSONPlanByRunID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	opts := otf.PlanFileOptions{Format: otf.PlanJSONFormat}

	s.getPlanFile(w, r, vars["id"], opts)
}

func (s *Server) getPlanFile(w http.ResponseWriter, r *http.Request, runID string, opts otf.PlanFileOptions) {
	json, err := s.RunService.GetPlanFile(r.Context(), runID, opts)
	if err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	if _, err := w.Write(json); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// RunJSONAPIObject converts a Run to a struct
// that can be marshalled into a JSON-API object
func RunJSONAPIObject(req *http.Request, r *otf.Run) *Run {
	result := &Run{
		ID: r.ID,
		Actions: &RunActions{
			IsCancelable:      r.IsCancelable(),
			IsConfirmable:     r.IsConfirmable(),
			IsForceCancelable: r.IsForceCancelable(),
			IsDiscardable:     r.IsDiscardable(),
		},
		CreatedAt:              r.CreatedAt,
		ForceCancelAvailableAt: r.ForceCancelAvailableAt(),
		HasChanges:             r.Plan.HasChanges(),
		IsDestroy:              r.IsDestroy,
		Message:                r.Message,
		Permissions: &otf.RunPermissions{
			CanForceCancel:  true,
			CanApply:        true,
			CanCancel:       true,
			CanDiscard:      true,
			CanForceExecute: true,
		},
		PositionInQueue: 0,
		Refresh:         r.Refresh,
		RefreshOnly:     r.RefreshOnly,
		ReplaceAddrs:    r.ReplaceAddrs,
		Source:          otf.DefaultConfigurationSource,
		Status:          r.Status,
		TargetAddrs:     r.TargetAddrs,

		// Relations
		Apply:                ApplyJSONAPIObject(req, r.Apply),
		ConfigurationVersion: ConfigurationVersionJSONAPIObject(r.ConfigurationVersion),
		Plan:                 PlanJSONAPIObject(req, r.Plan),
		Workspace:            WorkspaceJSONAPIObject(r.Workspace),

		// Hardcoded anonymous user until authorization is introduced
		CreatedBy: &User{
			ID:       otf.DefaultUserID,
			Username: otf.DefaultUsername,
		},
	}

	for k, v := range r.StatusTimestamps {
		if result.StatusTimestamps == nil {
			result.StatusTimestamps = &RunStatusTimestamps{}
		}
		switch otf.RunStatus(k) {
		case otf.RunPending:
			result.StatusTimestamps.PlanQueueableAt = &v
		case otf.RunPlanQueued:
			result.StatusTimestamps.PlanQueuedAt = &v
		case otf.RunPlanning:
			result.StatusTimestamps.PlanningAt = &v
		case otf.RunPlanned:
			result.StatusTimestamps.PlannedAt = &v
		case otf.RunPlannedAndFinished:
			result.StatusTimestamps.PlannedAndFinishedAt = &v
		case otf.RunApplyQueued:
			result.StatusTimestamps.ApplyQueuedAt = &v
		case otf.RunApplying:
			result.StatusTimestamps.ApplyingAt = &v
		case otf.RunApplied:
			result.StatusTimestamps.AppliedAt = &v
		case otf.RunErrored:
			result.StatusTimestamps.ErroredAt = &v
		case otf.RunCanceled:
			result.StatusTimestamps.CanceledAt = &v
		case otf.RunForceCanceled:
			result.StatusTimestamps.ForceCanceledAt = &v
		case otf.RunDiscarded:
			result.StatusTimestamps.DiscardedAt = &v
		}
	}

	return result
}

// RunListJSONAPIObject converts a RunList to
// a struct that can be marshalled into a JSON-API object
func RunListJSONAPIObject(req *http.Request, cvl *otf.RunList) *RunList {
	obj := &RunList{
		Pagination: cvl.Pagination,
	}
	for _, item := range cvl.Items {
		obj.Items = append(obj.Items, RunJSONAPIObject(req, item))
	}

	return obj
}
