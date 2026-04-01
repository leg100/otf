package runner

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/tfeapi"
)

type (
	API struct {
		*tfeapi.Responder
		Client apiClient
	}

	apiClient interface {
		CreateAgentToken(ctx context.Context, poolID resource.ID, opts CreateAgentTokenOptions) (*AgentToken, []byte, error)

		Register(ctx context.Context, opts RegisterRunnerOptions) (*RunnerMeta, error)
		awaitAllocatedJobs(ctx context.Context, agentID resource.ID) ([]*Job, error)
		updateStatus(ctx context.Context, agentID resource.ID, status RunnerStatus) error

		GetJob(ctx context.Context, jobID resource.ID) (*Job, error)
		startJob(ctx context.Context, jobID resource.ID) ([]byte, error)
		awaitJobSignal(ctx context.Context, jobID resource.ID) func() (jobSignal, error)
		finishJob(ctx context.Context, jobID resource.ID, opts finishJobOptions) error

		GenerateDynamicCredentialsToken(ctx context.Context, jobID resource.ID, audience string) ([]byte, error)
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

	generateDynamicCredentialsTokenParams struct {
		Audience string `json:"audience"`
	}
)

func (a *API) AddHandlers(r *mux.Router) {
	r.HandleFunc("/agents/register", a.registerAgent).Methods("POST")
	r.HandleFunc("/agents/status", a.updateAgentStatus).Methods("POST")
	r.HandleFunc("/agents/await-allocated-jobs", a.awaitAllocatedJobs).Methods("GET")
	r.HandleFunc("/jobs/start", a.startJob).Methods("POST")
	r.HandleFunc("/jobs/finish", a.finishJob).Methods("POST")
	r.HandleFunc("/jobs/{job_id}", a.getJob).Methods("GET")
	r.HandleFunc("/jobs/{job_id}/await-signal", a.awaitJobSignal).Methods("GET")
	r.HandleFunc("/jobs/{job_id}/dynamic-credentials", a.generateDynamicCredentialsToken).Methods("POST")

	// agent tokens
	r.HandleFunc("/agent-tokens/{pool_id}/create", a.createAgentToken).Methods("POST")
}

func (a *API) registerAgent(w http.ResponseWriter, r *http.Request) {
	var opts RegisterRunnerOptions
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

	agent, err := a.Client.Register(r.Context(), opts)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	a.Respond(w, r, agent, http.StatusCreated)
}

func (a *API) awaitAllocatedJobs(w http.ResponseWriter, r *http.Request) {
	// retrieve runner, which contains ID of calling agent
	runner, err := runnerFromContext(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	jobs, err := a.Client.awaitAllocatedJobs(r.Context(), runner.ID)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	a.Respond(w, r, jobs, http.StatusOK)
}

func (a *API) getJob(w http.ResponseWriter, r *http.Request) {
	jobID, err := decode.ID("job_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	job, err := a.Client.GetJob(r.Context(), jobID)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	a.Respond(w, r, job, http.StatusOK)
}

func (a *API) awaitJobSignal(w http.ResponseWriter, r *http.Request) {
	jobID, err := decode.ID("job_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	signal, err := a.Client.awaitJobSignal(r.Context(), jobID)()
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	a.Respond(w, r, signal, http.StatusOK)
}

// updateAgentStatus receives a status update from an agent
func (a *API) updateAgentStatus(w http.ResponseWriter, r *http.Request) {
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

	err = a.Client.updateStatus(r.Context(), runner.ID, params.Status)
	if err != nil {
		if errors.Is(err, ErrInvalidStateTransition) {
			tfeapi.Error(w, err)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
}

func (a *API) createAgentToken(w http.ResponseWriter, r *http.Request) {
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
	_, token, err := a.Client.CreateAgentToken(r.Context(), poolID, opts)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	w.Write(token)
}

func (a *API) startJob(w http.ResponseWriter, r *http.Request) {
	var params startJobParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		tfeapi.Error(w, err, tfeapi.WithStatus(http.StatusUnprocessableEntity))
		return
	}
	token, err := a.Client.startJob(r.Context(), params.JobID)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	w.Write(token)
}

func (a *API) finishJob(w http.ResponseWriter, r *http.Request) {
	var params finishJobParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		tfeapi.Error(w, err, tfeapi.WithStatus(http.StatusUnprocessableEntity))
		return
	}
	err := a.Client.finishJob(r.Context(), params.JobID, params.finishJobOptions)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
}

func (a *API) generateDynamicCredentialsToken(w http.ResponseWriter, r *http.Request) {
	jobID, err := decode.ID("job_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	var params generateDynamicCredentialsTokenParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		tfeapi.Error(w, err, tfeapi.WithStatus(http.StatusUnprocessableEntity))
		return
	}
	token, err := a.Client.GenerateDynamicCredentialsToken(
		r.Context(),
		jobID,
		params.Audience,
	)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	w.Write(token)
}
