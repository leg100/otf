package http

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/go-tfe"
	"github.com/leg100/jsonapi"
)

func (s *Server) CreateRun(w http.ResponseWriter, r *http.Request) {
	CreateObject(w, r, &tfe.RunCreateOptions{}, func(opts interface{}) (interface{}, error) {
		return s.RunService.CreateRun(opts.(*tfe.RunCreateOptions))
	})
}

func (s *Server) ApplyRun(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	opts := &tfe.RunApplyOptions{}
	if err := jsonapi.UnmarshalPayload(r.Body, opts); err != nil {
		ErrUnprocessable(w, err)
		return
	}

	if err := s.RunService.ApplyRun(vars["id"], opts); err != nil {
		ErrNotFound(w)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	w.Header().Set("Content-type", jsonapi.MediaType)
}

func (s *Server) GetRun(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	GetObject(w, r, func() (interface{}, error) {
		return s.RunService.GetRun(vars["id"])
	})
}

func (s *Server) ListRuns(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	var opts tfe.RunListOptions
	if err := decoder.Decode(&opts, r.URL.Query()); err != nil {
		ErrUnprocessable(w, fmt.Errorf("unable to decode query string: %w", err))
		return
	}

	SanitizeListOptions(&opts.ListOptions)

	ListObjects(w, r, func() (interface{}, error) {
		return s.RunService.ListRuns(vars["workspace_id"], opts)
	})
}

func (s *Server) DiscardRun(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	opts := &tfe.RunDiscardOptions{}
	if err := jsonapi.UnmarshalPayload(r.Body, opts); err != nil {
		ErrUnprocessable(w, err)
		return
	}

	if err := s.RunService.DiscardRun(vars["id"], opts); err != nil {
		ErrNotFound(w)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	w.Header().Set("Content-type", jsonapi.MediaType)
}

func (s *Server) CancelRun(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	opts := &tfe.RunCancelOptions{}
	if err := jsonapi.UnmarshalPayload(r.Body, opts); err != nil {
		ErrUnprocessable(w, err)
		return
	}

	if err := s.RunService.CancelRun(vars["id"], opts); err != nil {
		ErrNotFound(w)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	w.Header().Set("Content-type", jsonapi.MediaType)
}

func (s *Server) ForceCancelRun(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	opts := &tfe.RunForceCancelOptions{}
	if err := jsonapi.UnmarshalPayload(r.Body, opts); err != nil {
		ErrUnprocessable(w, err)
		return
	}

	if err := s.RunService.ForceCancelRun(vars["id"], opts); err != nil {
		ErrNotFound(w)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	w.Header().Set("Content-type", jsonapi.MediaType)
}
