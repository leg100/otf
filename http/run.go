package http

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/go-tfe"
	"github.com/leg100/jsonapi"
	"github.com/leg100/ots"
)

func (s *Server) CreateRun(w http.ResponseWriter, r *http.Request) {
	opts := tfe.RunCreateOptions{}
	if err := jsonapi.UnmarshalPayload(r.Body, &opts); err != nil {
		WriteError(w, http.StatusUnprocessableEntity, err)
		return
	}

	obj, err := s.RunService.Create(&opts)
	if err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	WriteResponse(w, r, s.RunJSONAPIObject(obj), WithCode(http.StatusCreated))
}

func (s *Server) GetRun(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	obj, err := s.RunService.Get(vars["id"])
	if err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	WriteResponse(w, r, s.RunJSONAPIObject(obj))
}

func (s *Server) ListRuns(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	var opts ots.RunListOptions
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

	WriteResponse(w, r, s.RunListJSONAPIObject(obj))
}

func (s *Server) ApplyRun(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	opts := &tfe.RunApplyOptions{}
	if err := jsonapi.UnmarshalPayload(r.Body, opts); err != nil {
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

	opts := &tfe.RunDiscardOptions{}
	if err := jsonapi.UnmarshalPayload(r.Body, opts); err != nil {
		WriteError(w, http.StatusUnprocessableEntity, err)
		return
	}

	err := s.RunService.Discard(vars["id"], opts)
	if err == ots.ErrRunDiscardNotAllowed {
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

	opts := &tfe.RunCancelOptions{}
	if err := jsonapi.UnmarshalPayload(r.Body, opts); err != nil {
		WriteError(w, http.StatusUnprocessableEntity, err)
		return
	}

	err := s.RunService.Cancel(vars["id"], opts)
	if err == ots.ErrRunCancelNotAllowed {
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

	opts := &tfe.RunForceCancelOptions{}
	if err := jsonapi.UnmarshalPayload(r.Body, opts); err != nil {
		WriteError(w, http.StatusUnprocessableEntity, err)
		return
	}

	err := s.RunService.ForceCancel(vars["id"], opts)
	if err == ots.ErrRunForceCancelNotAllowed {
		WriteError(w, http.StatusConflict, err)
		return
	} else if err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func (s *Server) GetRunPlanJSON(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	json, err := s.RunService.GetPlanJSON(vars["id"])
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
func (s *Server) RunJSONAPIObject(r *ots.Run) *tfe.Run {
	obj := &tfe.Run{
		ID:                     r.ID,
		Actions:                r.Actions(),
		CreatedAt:              r.CreatedAt,
		ForceCancelAvailableAt: r.ForceCancelAvailableAt,
		HasChanges:             r.Plan.HasChanges(),
		IsDestroy:              r.IsDestroy,
		Message:                r.Message,
		Permissions:            r.Permissions,
		PositionInQueue:        0,
		Refresh:                r.Refresh,
		RefreshOnly:            r.RefreshOnly,
		ReplaceAddrs:           r.ReplaceAddrs,
		Source:                 ots.DefaultConfigurationSource,
		Status:                 r.Status,
		StatusTimestamps:       r.StatusTimestamps,
		TargetAddrs:            r.TargetAddrs,

		// Relations
		Apply:                s.ApplyJSONAPIObject(r.Apply),
		ConfigurationVersion: s.ConfigurationVersionJSONAPIObject(r.ConfigurationVersion),
		Plan:                 s.PlanJSONAPIObject(r.Plan),
		Workspace:            s.WorkspaceJSONAPIObject(r.Workspace),

		// Hardcoded anonymous user until authorization is introduced
		CreatedBy: &tfe.User{
			ID:       ots.DefaultUserID,
			Username: ots.DefaultUsername,
		},
	}

	return obj
}

// RunListJSONAPIObject converts a RunList to
// a struct that can be marshalled into a JSON-API object
func (s *Server) RunListJSONAPIObject(cvl *ots.RunList) *tfe.RunList {
	obj := &tfe.RunList{
		Pagination: cvl.Pagination,
	}
	for _, item := range cvl.Items {
		obj.Items = append(obj.Items, s.RunJSONAPIObject(item))
	}

	return obj
}
