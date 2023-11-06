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
	"github.com/leg100/otf/internal/tokens"
)

type api struct {
	*service
	*tfeapi.Responder
}

func (a *api) addHandlers(r *mux.Router) {
	r = r.PathPrefix(otfapi.DefaultBasePath).Subrouter()
	r.HandleFunc("/agent/register", a.registerAgent).Methods("POST")
	r.HandleFunc("/agent/{agent_id}/jobs", a.getJobs).Methods("GET")
	r.HandleFunc("/agent/{agent_id}/status", a.updateStatus).Methods("POST")
}

func (a *api) registerAgent(w http.ResponseWriter, r *http.Request) {
	// middleware should have put agent token into context.
	token, err := tokens.AgentFromContext(r.Context())
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

	jobs, err := a.service.getAllocatedJobs(r.Context(), agentID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	a.Respond(w, r, jobs, http.StatusOK)
}

// updateStatus receives a status update from an agent, and optionally a job
// status update as well.
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
		Job    *jobParams
	}
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	if err := a.service.updateAgentStatus(r.Context(), agentID, params.Status); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if params.Job != nil {
		err = a.service.updateJobStatus(r.Context(), params.Job.JobSpec, params.Job.Status)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

// check agent_id has not been spoofed by checking it belongs to the pool of
// the token it has authenticated with.
func (a *api) spoofCheck(ctx context.Context, agentID string) error {
	token, err := tokens.AgentFromContext(ctx)
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
