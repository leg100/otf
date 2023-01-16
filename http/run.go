package http

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/leg100/jsonapi"
	"github.com/leg100/otf"
	"github.com/leg100/otf/http/decode"
	"github.com/leg100/otf/http/dto"
)

type planFileOptions struct {
	Format otf.PlanFormat `schema:"format,required"`
}

func (s *Server) CreateRun(w http.ResponseWriter, r *http.Request) {
	opts := dto.RunCreateOptions{}
	if err := jsonapi.UnmarshalPayload(r.Body, &opts); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err)
		return
	}
	if opts.Workspace == nil {
		writeError(w, http.StatusUnprocessableEntity, fmt.Errorf("missing workspace"))
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
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeResponse(w, r, &Run{run, r, s}, withCode(http.StatusCreated))
}

func (s *Server) startPhase(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	opts := otf.PhaseStartOptions{}
	if err := jsonapi.UnmarshalPayload(r.Body, &opts); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err)
		return
	}
	run, err := s.Application.StartPhase(
		r.Context(),
		vars["id"],
		otf.PhaseType(vars["phase"]),
		opts)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeResponse(w, r, &Run{run, r, s})
}

func (s *Server) finishPhase(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	opts := otf.PhaseFinishOptions{}
	if err := jsonapi.UnmarshalPayload(r.Body, &opts); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err)
		return
	}
	run, err := s.Application.FinishPhase(
		r.Context(),
		vars["id"],
		otf.PhaseType(vars["phase"]),
		opts)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeResponse(w, r, &Run{run, r, s})
}

func (s *Server) GetRun(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	run, err := s.Application.GetRun(r.Context(), vars["id"])
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeResponse(w, r, &Run{run, r, s})
}

func (s *Server) ListRuns(w http.ResponseWriter, r *http.Request) {
	s.listRuns(w, r, otf.RunListOptions{})
}

func (s *Server) GetRunsQueue(w http.ResponseWriter, r *http.Request) {
	s.listRuns(w, r, otf.RunListOptions{
		Statuses: []otf.RunStatus{otf.RunPlanQueued, otf.RunApplyQueued},
	})
}

func (s *Server) listRuns(w http.ResponseWriter, r *http.Request, opts otf.RunListOptions) {
	if err := decode.Query(&opts, r.URL.Query()); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err)
		return
	}
	if err := decode.Route(&opts, r); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err)
		return
	}
	rl, err := s.Application.ListRuns(r.Context(), opts)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeResponse(w, r, &RunList{rl, r, s})
}

func (s *Server) ApplyRun(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	opts := otf.RunApplyOptions{}
	if err := jsonapi.UnmarshalPayload(r.Body, &opts); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err)
		return
	}
	if err := s.Application.ApplyRun(r.Context(), vars["id"], opts); err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func (s *Server) DiscardRun(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	opts := otf.RunDiscardOptions{}
	if err := jsonapi.UnmarshalPayload(r.Body, &opts); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err)
		return
	}
	err := s.Application.DiscardRun(r.Context(), vars["id"], opts)
	if err == otf.ErrRunDiscardNotAllowed {
		writeError(w, http.StatusConflict, err)
		return
	} else if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func (s *Server) CancelRun(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	opts := otf.RunCancelOptions{}
	if err := jsonapi.UnmarshalPayload(r.Body, &opts); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err)
		return
	}
	err := s.Application.CancelRun(r.Context(), vars["id"], opts)
	if err == otf.ErrRunCancelNotAllowed {
		writeError(w, http.StatusConflict, err)
		return
	} else if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func (s *Server) ForceCancelRun(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	opts := otf.RunForceCancelOptions{}
	if err := jsonapi.UnmarshalPayload(r.Body, &opts); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err)
		return
	}
	err := s.Application.ForceCancelRun(r.Context(), vars["id"], opts)
	if err == otf.ErrRunForceCancelNotAllowed {
		writeError(w, http.StatusConflict, err)
		return
	} else if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func (s *Server) getPlanFile(w http.ResponseWriter, r *http.Request) {
	opts := planFileOptions{}
	if err := decode.Query(&opts, r.URL.Query()); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err)
		return
	}
	vars := mux.Vars(r)
	file, err := s.GetPlanFile(r.Context(), vars["run_id"], opts.Format)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	if _, err := w.Write(file); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) uploadPlanFile(w http.ResponseWriter, r *http.Request) {
	opts := planFileOptions{}
	if err := decode.Query(&opts, r.URL.Query()); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err)
		return
	}
	vars := mux.Vars(r)
	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, r.Body); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err)
		return
	}
	err := s.UploadPlanFile(r.Context(), vars["run_id"], buf.Bytes(), opts.Format)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func (s *Server) getLockFile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	file, err := s.GetLockFile(r.Context(), vars["run_id"])
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	if _, err := w.Write(file); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) uploadLockFile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, r.Body); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err)
		return
	}
	err := s.UploadLockFile(r.Context(), vars["run_id"], buf.Bytes())
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

type Run struct {
	*otf.Run
	req *http.Request
	*Server
}

// ToJSONAPI assembles a JSON-API DTO.
func (r *Run) ToJSONAPI() any {
	subject, err := otf.SubjectFromContext(r.req.Context())
	if err != nil {
		panic(err.Error())
	}
	perms, err := r.ListWorkspacePermissions(r.req.Context(), r.WorkspaceID())
	if err != nil {
		panic(err.Error())
	}
	policy := &otf.WorkspacePolicy{
		Organization: r.Organization(),
		WorkspaceID:  r.WorkspaceID(),
		Permissions:  perms,
	}

	obj := &dto.Run{
		ID: r.ID(),
		Actions: &dto.RunActions{
			IsCancelable:      r.Cancelable(),
			IsConfirmable:     r.Confirmable(),
			IsForceCancelable: r.ForceCancelAvailableAt() != nil,
			IsDiscardable:     r.Discardable(),
		},
		CreatedAt:              r.CreatedAt(),
		ExecutionMode:          string(r.ExecutionMode()),
		ForceCancelAvailableAt: r.ForceCancelAvailableAt(),
		HasChanges:             r.Plan().HasChanges(),
		IsDestroy:              r.IsDestroy(),
		Message:                r.Message(),
		Permissions: &dto.RunPermissions{
			CanDiscard:      subject.CanAccessWorkspace(otf.DiscardRunAction, policy),
			CanForceExecute: subject.CanAccessWorkspace(otf.ApplyRunAction, policy),
			CanForceCancel:  subject.CanAccessWorkspace(otf.CancelRunAction, policy),
			CanCancel:       subject.CanAccessWorkspace(otf.CancelRunAction, policy),
			CanApply:        subject.CanAccessWorkspace(otf.ApplyRunAction, policy),
		},
		PositionInQueue:  0,
		Refresh:          r.Refresh(),
		RefreshOnly:      r.RefreshOnly(),
		ReplaceAddrs:     r.ReplaceAddrs(),
		Source:           otf.DefaultConfigurationSource,
		Status:           string(r.Status()),
		StatusTimestamps: &dto.RunStatusTimestamps{},
		TargetAddrs:      r.TargetAddrs(),
		// Relations
		Apply: (&apply{r.Apply(), r.req, r.Server}).ToJSONAPI().(*dto.Apply),
		Plan:  (&plan{r.Plan(), r.req, r.Server}).ToJSONAPI().(*dto.Plan),
		// Hardcoded anonymous user until authorization is introduced
		CreatedBy: &dto.User{
			ID:       otf.DefaultUserID,
			Username: otf.DefaultUsername,
		},
		ConfigurationVersion: &dto.ConfigurationVersion{
			ID: r.ConfigurationVersionID(),
		},
		Workspace: &dto.Workspace{ID: r.WorkspaceID()},
	}

	// Support including related resources:
	//
	// https://developer.hashicorp.com/terraform/cloud-docs/api-docs/run#available-related-resources
	//
	// NOTE: limit support to workspace, since that's what the go-tfe tests
	// for, and we want to run the full barrage of go-tfe workspace tests
	// without error
	if includes := r.req.URL.Query().Get("include"); includes != "" {
		for _, inc := range strings.Split(includes, ",") {
			switch inc {
			case "workspace":
				ws, err := r.Application.GetWorkspace(r.req.Context(), r.WorkspaceID())
				if err != nil {
					panic(err.Error()) // throws HTTP500
				}
				obj.Workspace = (&Workspace{r.req, r.Application, ws}).ToJSONAPI().(*dto.Workspace)
			}
		}
	}

	for _, rst := range r.StatusTimestamps() {
		switch rst.Status {
		case otf.RunPending:
			obj.StatusTimestamps.PlanQueueableAt = &rst.Timestamp
		case otf.RunPlanQueued:
			obj.StatusTimestamps.PlanQueuedAt = &rst.Timestamp
		case otf.RunPlanning:
			obj.StatusTimestamps.PlanningAt = &rst.Timestamp
		case otf.RunPlanned:
			obj.StatusTimestamps.PlannedAt = &rst.Timestamp
		case otf.RunPlannedAndFinished:
			obj.StatusTimestamps.PlannedAndFinishedAt = &rst.Timestamp
		case otf.RunApplyQueued:
			obj.StatusTimestamps.ApplyQueuedAt = &rst.Timestamp
		case otf.RunApplying:
			obj.StatusTimestamps.ApplyingAt = &rst.Timestamp
		case otf.RunApplied:
			obj.StatusTimestamps.AppliedAt = &rst.Timestamp
		case otf.RunErrored:
			obj.StatusTimestamps.ErroredAt = &rst.Timestamp
		case otf.RunCanceled:
			obj.StatusTimestamps.CanceledAt = &rst.Timestamp
		case otf.RunForceCanceled:
			obj.StatusTimestamps.ForceCanceledAt = &rst.Timestamp
		case otf.RunDiscarded:
			obj.StatusTimestamps.DiscardedAt = &rst.Timestamp
		}
	}
	return obj
}

type RunList struct {
	*otf.RunList
	req *http.Request
	*Server
}

// ToJSONAPI assembles a JSON-API DTO.
func (l *RunList) ToJSONAPI() any {
	obj := &dto.RunList{
		Pagination: l.Pagination.ToJSONAPI(),
	}
	for _, item := range l.Items {
		obj.Items = append(obj.Items, (&Run{item, l.req, l.Server}).ToJSONAPI().(*dto.Run))
	}
	return obj
}
