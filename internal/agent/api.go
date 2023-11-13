package agent

import (
	"context"
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

func (a *api) addHandlers(r *mux.Router) {
	r = r.PathPrefix(otfapi.DefaultBasePath).Subrouter()

	// agents
	r.HandleFunc("/agent/register", a.registerAgent).Methods("POST")
	r.HandleFunc("/agent/{agent_id}/jobs", a.getJobs).Methods("GET")
	r.HandleFunc("/agent/{agent_id}/status", a.updateStatus).Methods("POST")

	// agent tokens
	r.HandleFunc("/agent-tokens/create", a.createAgentToken).Methods("POST")

	// job tokens
	r.HandleFunc("/tokens/job", a.createJobToken).Methods("POST")
}

func (a *api) registerAgent(w http.ResponseWriter, r *http.Request) {
	// middleware should have put agent token into context.
	token, err := AgentTokenFromContext(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
	}

	var params struct {
		Name        *string // optional name
		Concurrency int
		CurrentJobs []JobSpec `json:"current_jobs"`
	}
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	agent, err := a.service.registerAgent(r.Context(), registerAgentOptions{
		Name:        params.Name,
		Concurrency: params.Concurrency,
		CurrentJobs: params.CurrentJobs,
		IPAddress:   net.ParseIP(r.RemoteAddr),
		AgentPoolID: &token.AgentPoolID,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	a.Respond(w, r, agent, http.StatusCreated)
}

func (a *api) getJobs(w http.ResponseWriter, r *http.Request) {
	agentID, err := decode.Param("agent_id", r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	if err := a.spoofCheck(r.Context(), agentID); err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	jobs, err := a.service.getAgentJobs(r.Context(), agentID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	a.Respond(w, r, jobs, http.StatusOK)
}

// updateStatus receives a status update from an agent, including both the
// status of the agent itself and the status of its jobs.
func (a *api) updateStatus(w http.ResponseWriter, r *http.Request) {
	agentID, err := decode.Param("agent_id", r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	if err := a.spoofCheck(r.Context(), agentID); err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	type jobParams struct {
		JobSpec
		Status JobStatus
	}
	var params struct {
		Status AgentStatus
		Jobs   []jobParams
	}
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	if err := a.service.updateAgentStatus(r.Context(), agentID, params.Status); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	for _, job := range params.Jobs {
		err = a.service.updateJobStatus(r.Context(), job.JobSpec, job.Status)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func (a *api) createAgentToken(w http.ResponseWriter, r *http.Request) {
	var opts CreateAgentTokenOptions
	if err := json.NewDecoder(r.Body).Decode(&opts); err != nil {
		tfeapi.Error(w, err)
		return
	}
	token, err := a.CreateAgentToken(r.Context(), opts)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	w.Write(token)
}

func (a *api) createJobToken(w http.ResponseWriter, r *http.Request) {
	var spec JobSpec
	if err := json.NewDecoder(r.Body).Decode(&spec); err != nil {
		tfeapi.Error(w, err)
		return
	}
	token, err := a.service.createJobToken(spec)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	w.Write(token)
}

// check agent_id has not been spoofed by checking it belongs to the pool of
// the token it has authenticated with.
func (a *api) spoofCheck(ctx context.Context, agentID string) error {
	token, err := AgentTokenFromContext(ctx)
	if err != nil {
		return err
	}
	agent, err := a.service.getAgent(ctx, agentID)
	if err != nil {
		return err
	}
	if agent.AgentPoolID == nil || *agent.AgentPoolID != token.AgentPoolID {
		return errors.New("authentication token does not belong to specified agent_id")
	}
	return nil
}
