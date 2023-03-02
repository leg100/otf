package auth

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/leg100/otf/http/jsonapi"
)

// api provides handlers for json:api endpoints
type api struct {
	app service
}

func (h *api) addHandlers(r *mux.Router) {
	r = r.PathPrefix("/api/v2").Subrouter()

	r.Use(AuthenticateToken(h.app))

	// Agent token routes
	r.HandleFunc("/agent/details", h.getCurrentAgent).Methods("GET")
	r.HandleFunc("/agent/create", h.createAgentToken).Methods("POST")

	// Registry session routes
	r.HandleFunc("/organizations/{organization_name}/registry/sessions/create", h.createRegistrySession).Methods("POST")

	// User routes
	r.HandleFunc("/account/details", h.getCurrentUser).Methods("GET")
}

// User routes

func (h *api) getCurrentUser(w http.ResponseWriter, r *http.Request) {
	user, err := otf.UserFromContext(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	jsonapi.WriteResponse(w, r, &jsonapi.User{ID: user.ID, Username: user.Username})
}

// Registry session routes

func (h *api) createRegistrySession(w http.ResponseWriter, r *http.Request) {
	var opts jsonapi.RegistrySessionCreateOptions
	if err := jsonapi.UnmarshalPayload(r.Body, &opts); err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}

	session, err := h.app.CreateRegistrySession(r.Context(), opts.OrganizationName)
	if err != nil {
		jsonapi.Error(w, http.StatusNotFound, err)
		return
	}

	jsonapi.WriteResponse(w, r, session)
}

// Agent token routes

func (h *api) createAgentToken(w http.ResponseWriter, r *http.Request) {
	var opts jsonapi.AgentTokenCreateOptions
	if err := jsonapi.UnmarshalPayload(r.Body, &opts); err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}
	at, err := h.app.createAgentToken(r.Context(), otf.CreateAgentTokenOptions{
		Description:  opts.Description,
		Organization: opts.Organization,
	})
	if err != nil {
		jsonapi.Error(w, http.StatusNotFound, err)
		return
	}
	jsonapi.WriteResponse(w, r, &jsonapi.AgentToken{
		ID:           at.id,
		Token:        otf.String(at.token),
		Organization: at.organization,
	})
}

func (h *api) getCurrentAgent(w http.ResponseWriter, r *http.Request) {
	at, err := agentFromContext(r.Context())
	if err != nil {
		jsonapi.Error(w, http.StatusNotFound, err)
		return
	}
	jsonapi.WriteResponse(w, r, &jsonapi.AgentToken{
		ID:           at.id,
		Token:        nil, // deliberately omit token
		Organization: at.organization,
	})
}
