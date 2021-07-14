package http

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/go-tfe"
	"github.com/leg100/jsonapi"
	"github.com/leg100/ots"
)

func (h *Server) CreateWorkspace(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	opts := tfe.WorkspaceCreateOptions{}
	if err := jsonapi.UnmarshalPayload(r.Body, &opts); err != nil {
		WriteError(w, http.StatusUnprocessableEntity, err)
		return
	}

	if err := opts.Valid(); err != nil {
		WriteError(w, http.StatusUnprocessableEntity, err)
		return
	}

	obj, err := h.WorkspaceService.Create(vars["org"], &opts)
	if err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	WriteResponse(w, r, obj, WithCode(http.StatusCreated))
}

func (h *Server) GetWorkspace(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	obj, err := h.WorkspaceService.Get(vars["name"], vars["org"])
	if err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	WriteResponse(w, r, obj)
}

func (h *Server) GetWorkspaceByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	obj, err := h.WorkspaceService.GetByID(vars["id"])
	if err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	WriteResponse(w, r, obj)
}

func (h *Server) ListWorkspaces(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	var opts tfe.WorkspaceListOptions
	if err := DecodeQuery(&opts, r.URL.Query()); err != nil {
		WriteError(w, http.StatusUnprocessableEntity, err)
		return
	}

	obj, err := h.WorkspaceService.List(vars["org"], opts)
	if err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	WriteResponse(w, r, obj)
}

func (h *Server) UpdateWorkspace(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	opts := tfe.WorkspaceUpdateOptions{}
	if err := jsonapi.UnmarshalPayload(r.Body, &opts); err != nil {
		WriteError(w, http.StatusUnprocessableEntity, err)
		return
	}

	if err := opts.Valid(); err != nil {
		WriteError(w, http.StatusUnprocessableEntity, err)
		return
	}

	obj, err := h.WorkspaceService.Update(vars["name"], vars["org"], &opts)
	if err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	WriteResponse(w, r, obj)
}

func (h *Server) UpdateWorkspaceByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	opts := tfe.WorkspaceUpdateOptions{}
	if err := jsonapi.UnmarshalPayload(r.Body, &opts); err != nil {
		WriteError(w, http.StatusUnprocessableEntity, err)
		return
	}

	if err := opts.Valid(); err != nil {
		WriteError(w, http.StatusUnprocessableEntity, err)
		return
	}

	obj, err := h.WorkspaceService.UpdateByID(vars["id"], &opts)
	if err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	WriteResponse(w, r, obj)
}

func (h *Server) LockWorkspace(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	opts := tfe.WorkspaceLockOptions{}
	if err := jsonapi.UnmarshalPayload(r.Body, &opts); err != nil {
		WriteError(w, http.StatusUnprocessableEntity, err)
		return
	}

	obj, err := h.WorkspaceService.Lock(vars["id"], opts)
	if err == ots.ErrWorkspaceAlreadyLocked {
		WriteError(w, http.StatusConflict, err)
		return
	} else if err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	WriteResponse(w, r, obj)
}

func (h *Server) UnlockWorkspace(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	obj, err := h.WorkspaceService.Unlock(vars["id"])
	if err == ots.ErrWorkspaceAlreadyUnlocked {
		WriteError(w, http.StatusConflict, err)
		return
	} else if err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	WriteResponse(w, r, obj)
}

func (h *Server) DeleteWorkspace(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	if err := h.WorkspaceService.Delete(vars["name"], vars["org"]); err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Server) DeleteWorkspaceByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	if err := h.WorkspaceService.DeleteByID(vars["id"]); err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
