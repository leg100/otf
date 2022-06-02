package http

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/jsonapi"
	"github.com/leg100/otf"
	"github.com/leg100/otf/http/decode"
	"github.com/leg100/otf/http/dto"
)

func (s *Server) CreateRun(w http.ResponseWriter, r *http.Request) {
	opts := dto.RunCreateOptions{}
	if err := jsonapi.UnmarshalPayload(r.Body, &opts); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err)
		return
	}
	if opts.Workspace == nil {
		writeError(w, http.StatusUnprocessableEntity, fmt.Errorf("missing workspace"))
	}
	var configurationVersionID *string
	if opts.ConfigurationVersion != nil {
		configurationVersionID = &opts.ConfigurationVersion.ID
	}
	run, err := s.RunService().Create(r.Context(), otf.RunCreateOptions{
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
		writeError(w, http.StatusNotFound, err)
		return
	}
	assembler := run.NewJSONAPIAssembler(r, string(GetPlanLogsRoute), string(GetApplyLogsRoute))
	writeResponse(w, r, assembler, withCode(http.StatusCreated))
}

func (s *Server) GetRun(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	run, err := s.RunService().Get(r.Context(), vars["id"])
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	assembler := run.NewJSONAPIAssembler(r, string(GetPlanLogsRoute), string(GetApplyLogsRoute))
	writeResponse(w, r, assembler)
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
	obj, err := s.RunService().List(r.Context(), opts)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeResponse(w, r, RunListDTO(r, obj))
}

func (s *Server) ApplyRun(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	opts := otf.RunApplyOptions{}
	if err := jsonapi.UnmarshalPayload(r.Body, &opts); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err)
		return
	}
	if err := s.RunService().Apply(r.Context(), vars["id"], opts); err != nil {
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
	err := s.RunService().Discard(r.Context(), vars["id"], opts)
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
	err := s.RunService().Cancel(r.Context(), vars["id"], opts)
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
	err := s.RunService().ForceCancel(r.Context(), vars["id"], opts)
	if err == otf.ErrRunForceCancelNotAllowed {
		writeError(w, http.StatusConflict, err)
		return
	} else if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func (s *Server) GetPlanFile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	var opts PlanFileOptions
	if err := decode.Query(&opts, r.URL.Query()); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err)
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
		writeError(w, http.StatusNotFound, err)
		return
	}
	if _, err := w.Write(json); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// RunListDTO converts a RunList to a struct that can be marshalled into a
// JSON-API object
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
