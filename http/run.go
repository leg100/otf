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

type uploadPlanFileOptions struct {
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
	workspace := otf.WorkspaceSpec{ID: &opts.Workspace.ID}
	var configurationVersionID *string
	if opts.ConfigurationVersion != nil {
		configurationVersionID = &opts.ConfigurationVersion.ID
	}
	run, err := s.Application.CreateRun(r.Context(), workspace, otf.RunCreateOptions{
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
	run, err := s.Application.GetRun(r.Context(), vars["id"])
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeResponse(w, r, run)
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
	if opts.Include != nil {
		for _, include := range strings.Split(*opts.Include, ",") {
			if include == "workspace" {
				ws, err := s.Application.GetWorkspace(r.Context(), otf.WorkspaceSpec{
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
	opts := uploadPlanFileOptions{}
	if err := decode.Query(&opts, r.URL.Query()); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err)
		return
	}
	vars := mux.Vars(r)
	file, err := s.Application.GetPlanFile(r.Context(), vars["run_id"], opts.Format)
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
	opts := uploadPlanFileOptions{}
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
	err := s.Application.UploadPlanFile(r.Context(), vars["run_id"], buf.Bytes(), opts.Format)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func (s *Server) getLockFile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	file, err := s.Application.GetLockFile(r.Context(), vars["run_id"])
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
	err := s.Application.UploadLockFile(r.Context(), vars["run_id"], buf.Bytes())
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}
