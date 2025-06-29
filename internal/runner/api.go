package runner

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	otfapi "github.com/leg100/otf/internal/api"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/tfeapi"
)

type (
	api struct {
		*Service
		*tfeapi.Responder
	}

	updateAgentStatusParams struct {
		Status RunnerStatus `json:"status"`
	}

	startJobParams struct {
		JobID resource.TfeID `json:"job_id"`
	}

	finishJobParams struct {
		finishJobOptions

		JobID resource.TfeID `json:"job_id"`
	}
)

func (a *api) addHandlers(r *mux.Router) {
	r = r.PathPrefix(otfapi.DefaultBasePath).Subrouter()

	// agents
	r.HandleFunc("/agents/register", a.registerAgent).Methods("POST")
	r.HandleFunc("/agents/jobs", a.getJobs).Methods("GET")
	r.HandleFunc("/agents/status", a.updateAgentStatus).Methods("POST")
	r.HandleFunc("/agents/start", a.startJob).Methods("POST")
	r.HandleFunc("/agents/finish", a.finishJob).Methods("POST")
	r.HandleFunc("/agents/jobs/{job_id}/await-signal", a.awaitJobSignal).Methods("GET")

	// agent tokens
	r.HandleFunc("/agent-tokens/{pool_id}/create", a.createAgentToken).Methods("POST")
}

func (a *api) registerAgent(w http.ResponseWriter, r *http.Request) {
	var opts registerOptions
	if err := json.NewDecoder(r.Body).Decode(&opts); err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// determine ip address from connection source address
	ip, err := internal.ParseAddr(r.RemoteAddr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	opts.IPAddress = &ip

	agent, err := a.Service.register(r.Context(), opts)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	a.Respond(w, r, agent, http.StatusCreated)
}

func (a *api) getJobs(w http.ResponseWriter, r *http.Request) {
	// retrieve runner, which contains ID of calling agent
	runner, err := runnerFromContext(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	jobs, err := a.Service.awaitAllocatedJobs(r.Context(), runner.ID)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	a.Respond(w, r, jobs, http.StatusOK)
}

func (a *api) awaitJobSignal(w http.ResponseWriter, r *http.Request) {
	jobID, err := decode.ID("job_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	signal, err := a.Service.awaitJobSignal(r.Context(), jobID)()
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	a.Respond(w, r, signal, http.StatusOK)
}

// updateAgentStatus receives a status update from an agent
func (a *api) updateAgentStatus(w http.ResponseWriter, r *http.Request) {
	// retrieve runner, which contains ID of calling agent
	runner, err := runnerFromContext(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	var params updateAgentStatusParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		tfeapi.Error(w, err)
		return
	}

	err = a.Service.updateStatus(r.Context(), runner.ID, params.Status)
	if err != nil {
		if errors.Is(err, ErrInvalidStateTransition) {
			tfeapi.Error(w, err)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
}

func (a *api) createAgentToken(w http.ResponseWriter, r *http.Request) {
	poolID, err := decode.ID("pool_id", r)
	if err != nil {
		tfeapi.Error(w, err)
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
	var params startJobParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		tfeapi.Error(w, err, tfeapi.WithStatus(http.StatusUnprocessableEntity))
		return
	}
	token, err := a.Service.startJob(r.Context(), params.JobID)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	w.Write(token)
}

func (a *api) finishJob(w http.ResponseWriter, r *http.Request) {
	var params finishJobParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		tfeapi.Error(w, err, tfeapi.WithStatus(http.StatusUnprocessableEntity))
		return
	}
	err := a.Service.finishJob(r.Context(), params.JobID, params.finishJobOptions)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
}
