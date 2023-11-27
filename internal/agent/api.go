package agent

import (
	"encoding/json"
	"errors"
	"net"
	"net/http"

	"github.com/gorilla/mux"
	otfapi "github.com/leg100/otf/internal/api"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/tfeapi"
)

type api struct {
	*service
	*tfeapi.Responder
}

type updateAgentStatusParams struct {
	Status AgentStatus `json:"status"`
}

func (a *api) addHandlers(r *mux.Router) {
	r = r.PathPrefix(otfapi.DefaultBasePath).Subrouter()

	// agents
	r.HandleFunc("/agents/register", a.registerAgent).Methods("POST")
	r.HandleFunc("/agents/jobs", a.getJobs).Methods("GET")
	r.HandleFunc("/agents/status", a.updateStatus).Methods("POST")
	r.HandleFunc("/agents/start", a.startJob).Methods("POST")

	// agent tokens
	r.HandleFunc("/agent-tokens/{pool_id}/create", a.createAgentToken).Methods("POST")
}

func (a *api) registerAgent(w http.ResponseWriter, r *http.Request) {
	var opts registerAgentOptions
	if err := json.NewDecoder(r.Body).Decode(&opts); err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// determine ip address from connection source address
	opts.IPAddress = net.ParseIP(r.RemoteAddr)

	agent, err := a.service.registerAgent(r.Context(), opts)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	a.Respond(w, r, agent, http.StatusCreated)
}

func (a *api) getJobs(w http.ResponseWriter, r *http.Request) {
	// retrieve subject, which contains ID of calling agent
	subject, err := poolAgentFromContext(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	jobs, err := a.service.getAgentJobs(r.Context(), subject.agent.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	a.Respond(w, r, jobs, http.StatusOK)
}

// updateStatus receives a status update from an agent
func (a *api) updateStatus(w http.ResponseWriter, r *http.Request) {
	// retrieve subject, which contains ID of calling agent
	subject, err := poolAgentFromContext(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	var params updateAgentStatusParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	err = a.service.updateAgentStatus(r.Context(), subject.agent.ID, params.Status)
	if err != nil {
		if errors.Is(err, ErrInvalidAgentStateTransition) {
			tfeapi.Error(w, err)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
}

func (a *api) createAgentToken(w http.ResponseWriter, r *http.Request) {
	poolID, err := decode.Param("pool_id", r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	var opts CreateAgentTokenOptions
	if err := json.NewDecoder(r.Body).Decode(&opts); err != nil {
		tfeapi.Error(w, err)
		return
	}
	_, token, err := a.CreateAgentToken(r.Context(), poolID, opts)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	w.Write(token)
}

func (a *api) startJob(w http.ResponseWriter, r *http.Request) {
	var spec JobSpec
	if err := json.NewDecoder(r.Body).Decode(&spec); err != nil {
		tfeapi.Error(w, err)
		return
	}
	token, err := a.service.startJob(r.Context(), spec)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	w.Write(token)
}
