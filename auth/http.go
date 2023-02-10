package auth

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/leg100/otf/http/jsonapi"
)

type handlers struct {
	app app
}

func (h *handlers) AddHandlers(r *mux.Router) {
	// TODO: add auth token mw
	r.HandleFunc("/account/details", h.GetCurrentUser).Methods("GET")

	// Registry session routes
	r.HandleFunc("/organizations/{organization_name}/registry/sessions/create", h.create)

	// Agent token routes
	r.HandleFunc("/agent/details", h.GetCurrentAgent).Methods("GET")
	r.HandleFunc("/agent/create", h.CreateAgentToken).Methods("POST")
}

// User routes

func (h *handlers) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	user, err := otf.UserFromContext(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	jsonapi.WriteResponse(w, r, &jsonapi.User{ID: user.ID(), Username: user.Username()})
}

// Registry session routes

func (h *handlers) create(w http.ResponseWriter, r *http.Request) {
	var opts jsonapi.RegistrySessionCreateOptions
	if err := jsonapi.UnmarshalPayload(r.Body, &opts); err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}

	session, err := h.app.createRegistrySession(r.Context(), opts.OrganizationName)
	if err != nil {
		jsonapi.Error(w, http.StatusNotFound, err)
		return
	}

	jsonapi.WriteResponse(w, r, session)
}

// Agent token routes

func (h *handlers) CreateAgentToken(w http.ResponseWriter, r *http.Request) {
	var opts jsonapi.AgentTokenCreateOptions
	if err := jsonapi.UnmarshalPayload(r.Body, &opts); err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}
	at, err := h.app.createAgentToken(r.Context(), CreateAgentTokenOptions{
		Description:  opts.Description,
		Organization: opts.Organization,
	})
	if err != nil {
		jsonapi.Error(w, http.StatusNotFound, err)
		return
	}
	jsonapi.WriteResponse(w, r, &jsonapi.AgentToken{at.id, at.token})
}

func (h *handlers) GetCurrentAgent(w http.ResponseWriter, r *http.Request) {
	at, err := agentFromContext(r.Context())
	if err != nil {
		jsonapi.Error(w, http.StatusNotFound, err)
		return
	}
	jsonapi.WriteResponse(w, r, &jsonapi.AgentTokenInfo{at.id, at.organization})
}
