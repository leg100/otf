package http

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/go-tfe"
	"github.com/leg100/jsonapi"
)

func (h *Server) CreateOrganization(w http.ResponseWriter, r *http.Request) {
	opts := tfe.OrganizationCreateOptions{}

	if err := jsonapi.UnmarshalPayload(r.Body, &opts); err != nil {
		WriteError(w, http.StatusUnprocessableEntity, err)
		return
	}

	if err := opts.Valid(); err != nil {
		WriteError(w, http.StatusUnprocessableEntity, err)
		return
	}

	obj, err := h.OrganizationService.Create(&opts)
	if err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	WriteResponse(w, r, obj, WithCode(http.StatusCreated))
}

func (h *Server) GetOrganization(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	obj, err := h.OrganizationService.Get(vars["name"])
	if err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	WriteResponse(w, r, obj)
}

func (h *Server) ListOrganizations(w http.ResponseWriter, r *http.Request) {
	var opts tfe.OrganizationListOptions

	if err := DecodeQuery(&opts, r.URL.Query()); err != nil {
		WriteError(w, http.StatusUnprocessableEntity, err)
		return
	}

	obj, err := h.OrganizationService.List(opts)
	if err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	WriteResponse(w, r, obj)
}

func (h *Server) UpdateOrganization(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]

	opts := tfe.OrganizationUpdateOptions{}
	if err := jsonapi.UnmarshalPayload(r.Body, &opts); err != nil {
		WriteError(w, http.StatusUnprocessableEntity, err)
		return
	}

	obj, err := h.OrganizationService.Update(name, &opts)
	if err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	WriteResponse(w, r, obj)
}

func (h *Server) DeleteOrganization(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]

	if err := h.OrganizationService.Delete(name); err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Server) GetEntitlements(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]

	obj, err := h.OrganizationService.GetEntitlements(name)
	if err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	WriteResponse(w, r, obj)
}
