package http

import (
	"fmt"
	"net/http"

	"github.com/google/jsonapi"
	"github.com/gorilla/mux"
	"github.com/hashicorp/go-tfe"
	"github.com/leg100/ots"
)

func (h *Server) ListWorkspaces(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	var opts ots.WorkspaceListOptions
	if err := decoder.Decode(&opts, r.URL.Query()); err != nil {
		ErrUnprocessable(w, fmt.Errorf("unable to decode query string: %w", err))
		return
	}

	SanitizeListOptions(&opts.ListOptions)

	ListObjects(w, r, func() (interface{}, error) {
		return h.WorkspaceService.ListWorkspaces(vars["org"], opts)
	})
}

func (h *Server) GetWorkspace(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	GetObject(w, r, func() (interface{}, error) {
		return h.WorkspaceService.GetWorkspace(vars["name"], vars["org"])
	})
}

func (h *Server) GetWorkspaceByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	GetObject(w, r, func() (interface{}, error) {
		return h.WorkspaceService.GetWorkspaceByID(vars["id"])
	})
}

func (h *Server) CreateWorkspace(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	CreateObject(w, r, &ots.WorkspaceCreateOptions{}, func(opts interface{}) (interface{}, error) {
		return h.WorkspaceService.CreateWorkspace(vars["org"], opts.(*ots.WorkspaceCreateOptions))
	})
}

func (h *Server) UpdateWorkspace(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	UpdateObject(w, r, &tfe.WorkspaceUpdateOptions{}, func(opts interface{}) (interface{}, error) {
		return h.WorkspaceService.UpdateWorkspace(vars["name"], vars["org"], opts.(*tfe.WorkspaceUpdateOptions))
	})
}

func (h *Server) UpdateWorkspaceByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	UpdateObject(w, r, &tfe.WorkspaceUpdateOptions{}, func(opts interface{}) (interface{}, error) {
		return h.WorkspaceService.UpdateWorkspaceByID(vars["id"], opts.(*tfe.WorkspaceUpdateOptions))
	})
}

func (h *Server) DeleteWorkspace(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	if err := h.WorkspaceService.DeleteWorkspace(vars["org"], vars["name"]); err != nil {
		ErrNotFound(w)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Server) DeleteWorkspaceByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	if err := h.WorkspaceService.DeleteWorkspaceByID(vars["id"]); err != nil {
		ErrNotFound(w)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Server) LockWorkspace(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	opts := ots.WorkspaceLockOptions{}
	if err := jsonapi.UnmarshalPayload(r.Body, &opts); err != nil {
		ErrUnprocessable(w, err)
		return
	}

	obj, err := h.WorkspaceService.LockWorkspace(vars["id"], opts)
	if err != nil {
		ErrNotFound(w)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-type", jsonapi.MediaType)
	if err := jsonapi.MarshalPayloadWithoutIncluded(w, obj); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *Server) UnlockWorkspace(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	obj, err := h.WorkspaceService.UnlockWorkspace(vars["id"])
	if err != nil {
		ErrNotFound(w)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-type", jsonapi.MediaType)
	if err := jsonapi.MarshalPayloadWithoutIncluded(w, obj); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
