package http

import (
	"fmt"
	"net/http"
	"strings"

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
		return
	}
	workspace := otf.WorkspaceSpec{ID: &opts.Workspace.ID}
	var configurationVersionID *string
	if opts.ConfigurationVersion != nil {
		configurationVersionID = &opts.ConfigurationVersion.ID
	}
	run, err := s.RunService().Create(r.Context(), workspace, otf.RunCreateOptions{
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
	writeResponse(w, r, run, withCode(http.StatusCreated))
}

func (s *Server) GetRun(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	run, err := s.RunService().Get(r.Context(), vars["id"])
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeResponse(w, r, run)
}

func (s *Server) ListRuns(w http.ResponseWriter, r *http.Request) {
	s.listRuns(w, r, RunListOptions{})
}

func (s *Server) GetRunsQueue(w http.ResponseWriter, r *http.Request) {
	s.listRuns(w, r, RunListOptions{
		Statuses: []otf.RunStatus{otf.RunPlanQueued, otf.RunApplyQueued},
	})
}

func (s *Server) listRuns(w http.ResponseWriter, r *http.Request, opts RunListOptions) {
	if err := decode.Query(&opts, r.URL.Query()); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err)
		return
	}
	if err := decode.Route(&opts, r); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err)
		return
	}
	rl, err := s.RunService().List(r.Context(), otf.RunListOptions{
		ListOptions:      opts.ListOptions,
		Statuses:         opts.Statuses,
		WorkspaceID:      opts.WorkspaceID,
		OrganizationName: opts.OrganizationName,
	})
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	if opts.Include != nil {
		for _, include := range strings.Split(*opts.Include, ",") {
			if include == "workspace" {
				ws, err := s.WorkspaceService().Get(r.Context(), otf.WorkspaceSpec{
					ID: opts.WorkspaceID,
				})
				if err != nil {
					writeError(w, http.StatusNotFound, err)
					return
				}
				for _, run := range rl.Items {
					run.IncludeWorkspace(ws)
				}
			}
		}
	}
	writeResponse(w, r, rl)
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

// RunGetOptions are options for retrieving a single Run. Either ID or ApplyID
// or PlanID must be specfiied.
type RunGetOptions struct {
	// A list of relations to include. See available resources:
	// https://www.terraform.io/docs/cloud/api/run.html#available-related-resources
	Include *string `schema:"include"`
}

// RunListOptions are options for paginating and filtering a list of runs
type RunListOptions struct {
	otf.ListOptions
	// A list of relations to include. See available resources:
	// https://www.terraform.io/docs/cloud/api/run.html#available-related-resources
	Include *string `schema:"include"`
	// Filter by run statuses (with an implicit OR condition)
	Statuses []otf.RunStatus
	// Filter by workspace ID
	WorkspaceID *string `schema:"workspace_id"`
	// Filter by organization name
	OrganizationName *string `schema:"organization_name"`
}
