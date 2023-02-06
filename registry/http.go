package registry

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/http/jsonapi"
)

type handlers struct {
	app service
}

func (h *handlers) AddHandlers(r *mux.Router) {
	// Registry session routes
	r.HandleFunc("/organizations/{organization_name}/registry/sessions/create", h.create)
}

func (h *handlers) create(w http.ResponseWriter, r *http.Request) {
	opts := jsonapiCreateOptions{}
	if err := jsonapi.UnmarshalPayload(r.Body, &opts); err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}
	session, err := h.app.create(r.Context(), opts.OrganizationName)
	if err != nil {
		jsonapi.Error(w, http.StatusNotFound, err)
		return
	}
	jsonapi.WriteResponse(w, r, session)
}
