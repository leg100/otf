package http

import (
	"net/http"

	"github.com/google/jsonapi"
	"github.com/gorilla/mux"
	"github.com/hashicorp/go-tfe"
	"github.com/leg100/ots"
)

func (h *Server) ListOrganizations(w http.ResponseWriter, r *http.Request) {
	orgs, err := h.OrganizationService.ListOrganizations()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Header().Set("Content-type", jsonapi.MediaType)
	if err := jsonapi.MarshalPayload(w, orgs); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *Server) GetOrganization(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]

	org, err := h.OrganizationService.GetOrganization(name)
	if err != nil {
		ErrNotFound(w)
		return
	}

	w.Header().Set("Content-type", jsonapi.MediaType)
	if err := jsonapi.MarshalPayload(w, org); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *Server) CreateOrganization(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]

	opts := &tfe.OrganizationCreateOptions{}
	if err := jsonapi.UnmarshalPayload(r.Body, opts); err != nil {
		ErrUnprocessable(w, err)
		return
	}

	org, err := ots.NewOrganizationFromOptions(opts)
	if err != nil {
		ErrUnprocessable(w, err)
		return
	}

	org, err = h.OrganizationService.CreateOrganization(name, org)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		jsonapi.MarshalErrors(w, []*jsonapi.ErrorObject{{
			Status: "404",
			Title:  "org already exists",
		}})
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-type", jsonapi.MediaType)
	if err := jsonapi.MarshalPayload(w, org); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *Server) UpdateOrganization(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]

	opts := &tfe.OrganizationUpdateOptions{}
	if err := jsonapi.UnmarshalPayload(r.Body, opts); err != nil {
		ErrUnprocessable(w, err)
		return
	}

	org, err := h.OrganizationService.UpdateOrganization(name, opts)
	if err != nil {
		ErrNotFound(w, WithDetail(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-type", jsonapi.MediaType)
	if err := jsonapi.MarshalPayload(w, org); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *Server) DeleteOrganization(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]

	if err := h.OrganizationService.DeleteOrganization(name); err != nil {
		ErrNotFound(w)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Server) GetEntitlements(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]

	entitlements, err := h.OrganizationService.GetEntitlements(name)
	if err != nil {
		ErrNotFound(w)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-type", jsonapi.MediaType)
	if err := jsonapi.MarshalPayload(w, entitlements); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

type ErrOption func(*jsonapi.ErrorObject)

func WithDetail(detail string) ErrOption {
	return func(err *jsonapi.ErrorObject) {
		err.Detail = detail
	}
}

func ErrNotFound(w http.ResponseWriter, opts ...ErrOption) {
	err := &jsonapi.ErrorObject{
		Status: "404",
		Title:  "not found",
	}

	for _, o := range opts {
		o(err)
	}

	w.WriteHeader(http.StatusNotFound)
	jsonapi.MarshalErrors(w, []*jsonapi.ErrorObject{
		err,
	})
}

func ErrUnprocessable(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusUnprocessableEntity)
	jsonapi.MarshalErrors(w, []*jsonapi.ErrorObject{{
		Status: "422",
		Title:  "unable to process payload",
		Detail: err.Error(),
	}})
}
