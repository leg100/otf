package http

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/go-tfe"
	"github.com/leg100/jsonapi"
	"github.com/leg100/ots"
)

func (h *Server) ListWorkspaces(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	var opts tfe.WorkspaceListOptions
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
	opts := tfe.WorkspaceCreateOptions{}

	if err := jsonapi.UnmarshalPayload(r.Body, &opts); err != nil {
		ErrUnprocessable(w, err)
		return
	}

	if err := opts.Valid(); err != nil {
		ErrUnprocessable(w, err)
		return
	}

	org, err := h.WorkspaceService.CreateWorkspace(vars["org"], &opts)
	if err != nil {
		ErrNotFound(w)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-type", jsonapi.MediaType)
	if err := jsonapi.MarshalPayloadWithoutIncluded(w, org); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *Server) UpdateWorkspace(w http.ResponseWriter, r *http.Request) {
	opts := tfe.WorkspaceUpdateOptions{}
	vars := mux.Vars(r)

	if err := jsonapi.UnmarshalPayload(r.Body, &opts); err != nil {
		ErrUnprocessable(w, err)
		return
	}

	if err := opts.Valid(); err != nil {
		ErrUnprocessable(w, err)
		return
	}

	ws, err := h.WorkspaceService.UpdateWorkspace(vars["name"], vars["org"], &opts)
	if err != nil {
		ErrNotFound(w)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-type", jsonapi.MediaType)
	if err := jsonapi.MarshalPayloadWithoutIncluded(w, ws); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *Server) UpdateWorkspaceByID(w http.ResponseWriter, r *http.Request) {
	opts := tfe.WorkspaceUpdateOptions{}
	vars := mux.Vars(r)

	if err := jsonapi.UnmarshalPayload(r.Body, &opts); err != nil {
		ErrUnprocessable(w, err)
		return
	}

	if err := opts.Valid(); err != nil {
		ErrUnprocessable(w, err)
		return
	}

	ws, err := h.WorkspaceService.UpdateWorkspaceByID(vars["id"], &opts)
	if err != nil {
		ErrNotFound(w)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-type", jsonapi.MediaType)
	if err := jsonapi.MarshalPayloadWithoutIncluded(w, ws); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *Server) DeleteWorkspace(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	if err := h.WorkspaceService.DeleteWorkspace(vars["name"], vars["org"]); err != nil {
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

	opts := tfe.WorkspaceLockOptions{}
	if err := jsonapi.UnmarshalPayload(r.Body, &opts); err != nil {
		ErrUnprocessable(w, err)
		return
	}

	obj, err := h.WorkspaceService.LockWorkspace(vars["id"], opts)
	if err == ots.ErrWorkspaceAlreadyLocked {
		WriteError(w, http.StatusConflict)
		return
	} else if err != nil {
		WriteError(w, http.StatusNotFound)
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
	if err == ots.ErrWorkspaceAlreadyUnlocked {
		WriteError(w, http.StatusConflict)
		return
	} else if err != nil {
		WriteError(w, http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-type", jsonapi.MediaType)
	if err := jsonapi.MarshalPayloadWithoutIncluded(w, obj); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
