package auth

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/leg100/otf/http/jsonapi"
)

type handlers struct {
	app appService
}

func (h *handlers) AddHandlers(r *mux.Router) {
	// TODO: add auth token mw
	r.HandleFunc("/account/details", h.GetCurrentUser).Methods("GET")
}

func (h *handlers) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	user, err := otf.UserFromContext(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	jsonapi.WriteResponse(w, r, &User{user})
}
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
package agenttoken

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/leg100/otf/http/jsonapi"
)

type handlers struct {
	app appService
}

func (h *handlers) AddHandlers(r *mux.Router) {
	r.HandleFunc("/agent/details", h.GetCurrentAgent).Methods("GET")
	r.HandleFunc("/agent/create", h.CreateAgentToken).Methods("POST")
}

func (h *handlers) CreateAgentToken(w http.ResponseWriter, r *http.Request) {
	opts := jsonapiCreateOptions{}
	if err := jsonapi.UnmarshalPayload(r.Body, &opts); err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}
	at, err := h.Application.CreateAgentToken(r.Context(), otf.CreateAgentTokenOptions{
		Description:  opts.Description,
		Organization: opts.Organization,
	})
	if err != nil {
		jsonapi.Error(w, http.StatusNotFound, err)
		return
	}
	jsonapi.WriteResponse(w, r, &agentToken{at, true})
}

func (h *handlers) GetCurrentAgent(w http.ResponseWriter, r *http.Request) {
	agent, err := otf.AgentFromContext(r.Context())
	if err != nil {
		jsonapi.Error(w, http.StatusNotFound, err)
		return
	}
	jsonapi.WriteResponse(w, r, &agentToken{agent, false})
}

type AgentToken struct {
	*otf.AgentToken
	revealToken bool // toggle send auth token over wire
}

// ToJSONAPI assembles a JSON-API DTO.
func (t *agentToken) ToJSONAPI() any {
	json := jsonapi.AgentToken{
		ID:           t.ID(),
		Organization: t.Organization(),
	}
	if t.revealToken {
		json.Token = t.Token()
	}
	return &json
}
