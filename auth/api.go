package auth

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	otfhttp "github.com/leg100/otf/http"
	"github.com/leg100/otf/http/decode"
	"github.com/leg100/otf/http/jsonapi"
)

// api provides handlers for json:api endpoints
type api struct {
	svc AuthService
}

func (h *api) addHandlers(r *mux.Router) {
	r = otfhttp.APIRouter(r)

	// Agent token routes
	r.HandleFunc("/agent/details", h.getCurrentAgent).Methods("GET")
	r.HandleFunc("/agent/create", h.createAgentToken).Methods("POST")

	// Registry session routes
	r.HandleFunc("/registry/sessions/create", h.createRegistrySession).Methods("POST")

	// User routes
	r.HandleFunc("/account/details", h.getCurrentUser).Methods("GET")
	r.HandleFunc("/admin/users", h.createUser).Methods("POST")
	r.HandleFunc("/admin/users/{username}", h.deleteUser).Methods("DELETE")
	r.HandleFunc("/teams/{team_id}/memberships/{username}", h.addTeamMembership).Methods("POST")
	r.HandleFunc("/teams/{team_id}/memberships/{username}", h.removeTeamMembership).Methods("DELETE")

	// Team routes
	r.HandleFunc("/organizations/{organization_name}/teams", h.createTeam).Methods("POST")
	r.HandleFunc("/organizations/{organization_name}/teams/{team_name}", h.getTeam).Methods("GET")
	r.HandleFunc("/teams/{team_id}", h.deleteTeam).Methods("DELETE")
}

// User routes

func (h *api) createUser(w http.ResponseWriter, r *http.Request) {
	var params jsonapi.CreateUserOptions
	if err := jsonapi.UnmarshalPayload(r.Body, &params); err != nil {
		jsonapi.Error(w, err)
		return
	}

	user, err := h.svc.CreateUser(r.Context(), *params.Username)
	if err != nil {
		jsonapi.Error(w, err)
		return
	}

	jsonapi.WriteResponse(w, r, &jsonapi.User{ID: user.ID, Username: user.Username}, jsonapi.WithCode(http.StatusCreated))
}

func (h *api) deleteUser(w http.ResponseWriter, r *http.Request) {
	username, err := decode.Param("username", r)
	if err != nil {
		jsonapi.Error(w, err)
		return
	}

	if err := h.svc.DeleteUser(r.Context(), username); err != nil {
		jsonapi.Error(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *api) addTeamMembership(w http.ResponseWriter, r *http.Request) {
	var params TeamMembershipOptions
	if err := decode.Route(&params, r); err != nil {
		jsonapi.Error(w, err)
		return
	}

	if err := h.svc.AddTeamMembership(r.Context(), params); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *api) removeTeamMembership(w http.ResponseWriter, r *http.Request) {
	var params TeamMembershipOptions
	if err := decode.Route(&params, r); err != nil {
		jsonapi.Error(w, err)
		return
	}

	if err := h.svc.RemoveTeamMembership(r.Context(), params); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *api) getCurrentUser(w http.ResponseWriter, r *http.Request) {
	user, err := UserFromContext(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	jsonapi.WriteResponse(w, r, &jsonapi.User{ID: user.ID, Username: user.Username})
}

// Team routes

func (h *api) createTeam(w http.ResponseWriter, r *http.Request) {
	var params jsonapi.CreateTeamOptions
	if err := decode.Route(&params, r); err != nil {
		jsonapi.Error(w, err)
		return
	}
	if err := jsonapi.UnmarshalPayload(r.Body, &params); err != nil {
		jsonapi.Error(w, err)
		return
	}

	team, err := h.svc.CreateTeam(r.Context(), CreateTeamOptions{
		Name:         *params.Name,
		Organization: *params.Organization,
	})
	if err != nil {
		jsonapi.Error(w, err)
		return
	}

	jsonapi.WriteResponse(w, r,
		&jsonapi.Team{ID: team.ID, Name: team.Name},
		jsonapi.WithCode(http.StatusCreated))
}

func (h *api) getTeam(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Organization *string `schema:"organization_name,required"`
		Name         *string `schema:"team_name,required"`
	}
	if err := decode.All(&params, r); err != nil {
		jsonapi.Error(w, err)
		return
	}

	team, err := h.svc.GetTeam(r.Context(), *params.Organization, *params.Name)
	if err != nil {
		jsonapi.Error(w, err)
		return
	}

	jsonapi.WriteResponse(w, r, &jsonapi.Team{ID: team.ID, Name: team.Name})
}

func (h *api) deleteTeam(w http.ResponseWriter, r *http.Request) {
	id, err := decode.Param("team_id", r)
	if err != nil {
		jsonapi.Error(w, err)
		return
	}

	if err := h.svc.DeleteTeam(r.Context(), id); err != nil {
		jsonapi.Error(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Registry session routes

func (h *api) createRegistrySession(w http.ResponseWriter, r *http.Request) {
	var opts jsonapi.RegistrySessionCreateOptions
	if err := jsonapi.UnmarshalPayload(r.Body, &opts); err != nil {
		jsonapi.Error(w, err)
		return
	}

	session, err := h.svc.CreateRegistrySession(r.Context(), CreateRegistrySessionOptions{
		Organization: opts.Organization,
	})
	if err != nil {
		jsonapi.Error(w, err)
		return
	}

	jsonapi.WriteResponse(w, r, &jsonapi.RegistrySession{
		Token:            session.Token,
		OrganizationName: session.Organization,
	})
}

// Agent token routes

func (h *api) createAgentToken(w http.ResponseWriter, r *http.Request) {
	var opts jsonapi.AgentTokenCreateOptions
	if err := jsonapi.UnmarshalPayload(r.Body, &opts); err != nil {
		jsonapi.Error(w, err)
		return
	}
	at, err := h.svc.CreateAgentToken(r.Context(), CreateAgentTokenOptions{
		Description:  opts.Description,
		Organization: opts.Organization,
	})
	if err != nil {
		jsonapi.Error(w, err)
		return
	}
	jsonapi.WriteResponse(w, r, &jsonapi.AgentToken{
		ID:           at.ID,
		Token:        otf.String(at.Token),
		Organization: at.Organization,
	})
}

func (h *api) getCurrentAgent(w http.ResponseWriter, r *http.Request) {
	at, err := agentFromContext(r.Context())
	if err != nil {
		jsonapi.Error(w, err)
		return
	}
	jsonapi.WriteResponse(w, r, &jsonapi.AgentToken{
		ID:           at.ID,
		Token:        nil, // deliberately omit token
		Organization: at.Organization,
	})
}
