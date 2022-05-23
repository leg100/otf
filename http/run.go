package http

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/jsonapi"
	"github.com/leg100/otf"
	"github.com/leg100/otf/http/dto"
)

func (s *Server) CreateRun(w http.ResponseWriter, r *http.Request) {
	opts := dto.RunCreateOptions{}
	if err := jsonapi.UnmarshalPayload(r.Body, &opts); err != nil {
		WriteError(w, http.StatusUnprocessableEntity, err)
		return
	}
	if opts.Workspace == nil {
		WriteError(w, http.StatusUnprocessableEntity, fmt.Errorf("missing workspace"))
	}
	var configurationVersionID *string
	if opts.ConfigurationVersion != nil {
		configurationVersionID = &opts.ConfigurationVersion.ID
	}
	obj, err := s.RunService().Create(r.Context(), otf.RunCreateOptions{
		IsDestroy:              opts.IsDestroy,
		Refresh:                opts.Refresh,
		RefreshOnly:            opts.RefreshOnly,
		Message:                opts.Message,
		ConfigurationVersionID: configurationVersionID,
		WorkspaceID:            opts.Workspace.ID,
		TargetAddrs:            opts.TargetAddrs,
		ReplaceAddrs:           opts.ReplaceAddrs,
	})
	if err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}
	WriteResponse(w, r, RunDTO(r, obj), WithCode(http.StatusCreated))
}

func (s *Server) GetRun(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	obj, err := s.RunService().Get(context.Background(), vars["id"])
	if err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	WriteResponse(w, r, RunDTO(r, obj))
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

	obj, err := s.RunService().List(context.Background(), opts)
	if err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	WriteResponse(w, r, RunListDTO(r, obj))
}

func (s *Server) UploadPlanFile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, r.Body); err != nil {
		WriteError(w, http.StatusUnprocessableEntity, err)
		return
	}

	var opts PlanFileOptions

	if err := DecodeQuery(&opts, r.URL.Query()); err != nil {
		WriteError(w, http.StatusUnprocessableEntity, err)
		return
	}

	if err := s.RunService().UploadPlanFile(r.Context(), vars["id"], buf.Bytes(), opts.Format); err != nil {
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

	if err := s.RunService().Apply(context.Background(), vars["id"], opts); err != nil {
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

	err := s.RunService().Discard(context.Background(), vars["id"], opts)
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

	err := s.RunService().Cancel(context.Background(), vars["id"], opts)
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

	err := s.RunService().ForceCancel(context.Background(), vars["id"], opts)
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
	id := vars["id"]

	var opts PlanFileOptions

	if err := DecodeQuery(&opts, r.URL.Query()); err != nil {
		WriteError(w, http.StatusUnprocessableEntity, err)
		return
	}

	s.getPlanFile(w, r, otf.RunGetOptions{ID: &id}, opts)
}

func (s *Server) GetJSONPlanByRunID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	opts := PlanFileOptions{Format: otf.PlanFormatJSON}

	s.getPlanFile(w, r, otf.RunGetOptions{ID: &id}, opts)
}

func (s *Server) getPlanFile(w http.ResponseWriter, r *http.Request, spec otf.RunGetOptions, opts PlanFileOptions) {
	json, err := s.RunService().GetPlanFile(r.Context(), spec, opts.Format)
	if err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	if _, err := w.Write(json); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// RunDTO converts a Run to a struct
// that can be marshalled into a JSON-API object
func RunDTO(req *http.Request, r *otf.Run) *dto.Run {
	result := &dto.Run{
		ID: r.ID(),
		Actions: &dto.RunActions{
			IsCancelable:      r.IsCancelable(),
			IsConfirmable:     r.IsConfirmable(),
			IsForceCancelable: r.IsForceCancelable(),
			IsDiscardable:     r.IsDiscardable(),
		},
		CreatedAt:              r.CreatedAt,
		ForceCancelAvailableAt: r.ForceCancelAvailableAt(),
		HasChanges:             r.Plan.HasChanges(),
		IsDestroy:              r.IsDestroy(),
		Message:                r.Message(),
		Permissions: &dto.RunPermissions{
			CanForceCancel:  true,
			CanApply:        true,
			CanCancel:       true,
			CanDiscard:      true,
			CanForceExecute: true,
		},
		PositionInQueue: 0,
		Refresh:         r.Refresh(),
		RefreshOnly:     r.RefreshOnly(),
		ReplaceAddrs:    r.ReplaceAddrs(),
		Source:          otf.DefaultConfigurationSource,
		Status:          string(r.Status()),
		TargetAddrs:     r.TargetAddrs(),

		// Relations
		Apply:                ApplyDTO(req, r.Apply),
		ConfigurationVersion: ConfigurationVersionDTO(r.ConfigurationVersion),
		Plan:                 PlanDTO(req, r.Plan),
		Workspace:            WorkspaceDTO(r.Workspace),

		// Hardcoded anonymous user until authorization is introduced
		CreatedBy: &dto.User{
			ID:       otf.DefaultUserID,
			Username: otf.DefaultUsername,
		},
	}

	for _, rst := range r.StatusTimestamps() {
		if result.StatusTimestamps == nil {
			result.StatusTimestamps = &dto.RunStatusTimestamps{}
		}
		switch rst.Status {
		case otf.RunPending:
			result.StatusTimestamps.PlanQueueableAt = &rst.Timestamp
		case otf.RunPlanQueued:
			result.StatusTimestamps.PlanQueuedAt = &rst.Timestamp
		case otf.RunPlanning:
			result.StatusTimestamps.PlanningAt = &rst.Timestamp
		case otf.RunPlanned:
			result.StatusTimestamps.PlannedAt = &rst.Timestamp
		case otf.RunPlannedAndFinished:
			result.StatusTimestamps.PlannedAndFinishedAt = &rst.Timestamp
		case otf.RunApplyQueued:
			result.StatusTimestamps.ApplyQueuedAt = &rst.Timestamp
		case otf.RunApplying:
			result.StatusTimestamps.ApplyingAt = &rst.Timestamp
		case otf.RunApplied:
			result.StatusTimestamps.AppliedAt = &rst.Timestamp
		case otf.RunErrored:
			result.StatusTimestamps.ErroredAt = &rst.Timestamp
		case otf.RunCanceled:
			result.StatusTimestamps.CanceledAt = &rst.Timestamp
		case otf.RunForceCanceled:
			result.StatusTimestamps.ForceCanceledAt = &rst.Timestamp
		case otf.RunDiscarded:
			result.StatusTimestamps.DiscardedAt = &rst.Timestamp
		}
	}

	return result
}

// RunListDTO converts a RunList to
// a struct that can be marshalled into a JSON-API object
func RunListDTO(req *http.Request, l *otf.RunList) *dto.RunList {
	pagination := dto.Pagination(*l.Pagination)
	obj := &dto.RunList{
		Pagination: &pagination,
	}
	for _, item := range l.Items {
		obj.Items = append(obj.Items, RunDTO(req, item))
	}

	return obj
}
