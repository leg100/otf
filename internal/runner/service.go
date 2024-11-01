package runner

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	otfhttp "github.com/leg100/otf/internal/http"
	"github.com/leg100/otf/internal/http/html"
	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/pubsub"
	"github.com/leg100/otf/internal/rbac"
	otfrun "github.com/leg100/otf/internal/run"
	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/sql/sqlc"
	"github.com/leg100/otf/internal/tfeapi"
	"github.com/leg100/otf/internal/tokens"
	"github.com/leg100/otf/internal/workspace"
)

var (
	ErrInvalidStateTransition   = errors.New("invalid runner state transition")
	ErrUnauthorizedRegistration = errors.New("unauthorized runner registration")
)

type (
	Service struct {
		logr.Logger

		organization internal.Authorizer

		tfeapi       *tfe
		api          *api
		web          *webHandlers
		poolBroker   pubsub.SubscriptionService[*Pool]
		runnerBroker pubsub.SubscriptionService[*RunnerMeta]
		jobBroker    pubsub.SubscriptionService[*Job]
		phases       phaseClient

		db *db
		*tokenFactory
	}

	ServiceOptions struct {
		logr.Logger
		*sql.DB
		*sql.Listener
		html.Renderer
		*tfeapi.Responder

		RunService       *otfrun.Service
		WorkspaceService *workspace.Service
		TokensService    *tokens.Service
	}

	phaseClient interface {
		StartPhase(ctx context.Context, runID string, phase internal.PhaseType, _ otfrun.PhaseStartOptions) (*otfrun.Run, error)
		FinishPhase(ctx context.Context, runID string, phase internal.PhaseType, opts otfrun.PhaseFinishOptions) (*otfrun.Run, error)
		Cancel(ctx context.Context, runID string) error
	}
)

func NewService(opts ServiceOptions) *Service {
	svc := &Service{
		Logger:       opts.Logger,
		db:           &db{DB: opts.DB},
		organization: &organization.Authorizer{Logger: opts.Logger},
		tokenFactory: &tokenFactory{
			tokens: opts.TokensService,
		},
		phases: opts.RunService,
	}
	svc.tfeapi = &tfe{
		Service:   svc,
		Responder: opts.Responder,
	}
	svc.api = &api{
		Service:   svc,
		Responder: opts.Responder,
	}
	svc.web = &webHandlers{
		Renderer:   opts.Renderer,
		logger:     opts.Logger,
		svc:        svc,
		workspaces: opts.WorkspaceService,
	}
	svc.poolBroker = pubsub.NewBroker(
		opts.Logger,
		opts.Listener,
		"agent_pools",
		func(ctx context.Context, id string, action sql.Action) (*Pool, error) {
			if action == sql.DeleteAction {
				return &Pool{ID: id}, nil
			}
			return svc.db.getPool(ctx, id)
		},
	)
	svc.runnerBroker = pubsub.NewBroker(
		opts.Logger,
		opts.Listener,
		"runners",
		func(ctx context.Context, id string, action sql.Action) (*RunnerMeta, error) {
			if action == sql.DeleteAction {
				return &RunnerMeta{ID: id}, nil
			}
			return svc.db.get(ctx, id)
		},
	)
	svc.jobBroker = pubsub.NewBroker(
		opts.Logger,
		opts.Listener,
		"jobs",
		func(ctx context.Context, specStr string, action sql.Action) (*Job, error) {
			spec, err := jobSpecFromString(specStr)
			if err != nil {
				return nil, err
			}
			if action == sql.DeleteAction {
				return &Job{Spec: spec}, nil
			}
			return svc.db.getJob(ctx, spec)
		},
	)
	// Register with auth middleware the agent token kind and a means of
	// retrieving the appropriate runner corresponding to the agent token ID
	opts.TokensService.RegisterKind(AgentTokenKind, func(ctx context.Context, tokenID string) (internal.Subject, error) {
		pool, err := svc.db.getPoolByTokenID(ctx, tokenID)
		if err != nil {
			return nil, err
		}
		// if the runner has registered then it should be sending its ID in an
		// http header
		headers, err := otfhttp.HeadersFromContext(ctx)
		if err != nil {
			return nil, err
		}
		if runnerID := headers.Get(runnerIDHeader); runnerID != "" {
			runner, err := svc.getRunner(ctx, runnerID)
			if err != nil {
				return nil, fmt.Errorf("retrieving runner corresponding to ID found in http header: %w", err)
			}
			return runner, nil
		}
		// Runner hasn't registered yet
		//
		// TODO: create constructor for constructing unregistered agent.
		return &RunnerMeta{AgentPoolID: &pool.ID}, nil
	})
	// create jobs when a plan or apply is enqueued
	opts.RunService.AfterEnqueuePlan(svc.createJob)
	opts.RunService.AfterEnqueueApply(svc.createJob)
	// cancel job when a run is canceled
	opts.RunService.AfterCancelRun(svc.cancelJob)
	// cancel job when a run is forceably canceled
	opts.RunService.AfterForceCancelRun(svc.cancelJob)
	// check whether a workspace is being created or updated and configured to
	// use an agent pool, and if so, check that it is allowed to use the pool.
	opts.WorkspaceService.BeforeCreateWorkspace(svc.checkWorkspacePoolAccess)
	opts.WorkspaceService.BeforeUpdateWorkspace(svc.checkWorkspacePoolAccess)
	// Register with auth middleware the job token and a means of
	// retrieving Job corresponding to token.
	opts.TokensService.RegisterKind(JobTokenKind, func(ctx context.Context, jobspecString string) (internal.Subject, error) {
		spec, err := jobSpecFromString(jobspecString)
		if err != nil {
			return nil, err
		}
		return svc.getJob(ctx, spec)
	})
	return svc
}

func (s *Service) AddHandlers(r *mux.Router) {
	s.tfeapi.addHandlers(r)
	s.api.addHandlers(r)
	s.web.addHandlers(r)
}

func (s *Service) NewAllocator(logger logr.Logger) *allocator {
	return &allocator{
		Logger: logger,
		client: s,
	}
}

func (s *Service) NewManager() *manager { return newManager(s) }

func (s *Service) WatchAgentPools(ctx context.Context) (<-chan pubsub.Event[*Pool], func()) {
	return s.poolBroker.Subscribe(ctx)
}

func (s *Service) WatchRunners(ctx context.Context) (<-chan pubsub.Event[*RunnerMeta], func()) {
	return s.runnerBroker.Subscribe(ctx)
}

func (s *Service) WatchJobs(ctx context.Context) (<-chan pubsub.Event[*Job], func()) {
	return s.jobBroker.Subscribe(ctx)
}

func (s *Service) register(ctx context.Context, opts registerOptions) (*RunnerMeta, error) {
	meta, err := func() (*RunnerMeta, error) {
		if err := authorizeRunner(ctx, ""); err != nil {
			return nil, ErrUnauthorizedRegistration
		}
		meta, err := register(opts)
		if err != nil {
			return nil, err
		}
		if err := s.db.create(ctx, meta); err != nil {
			return nil, err
		}
		return meta, nil
	}()
	if err != nil {
		s.Error(err, "registering runner")
		return nil, err
	}
	s.V(0).Info("registered runner", "runner", meta)
	return meta, nil
}

func (s *Service) getRunner(ctx context.Context, runnerID string) (*RunnerMeta, error) {
	return s.db.get(ctx, runnerID)
}

func (s *Service) updateStatus(ctx context.Context, runnerID string, to RunnerStatus) error {
	// only these subjects may call this endpoint:
	// (a) the manager, or
	// (b) an runner with an ID matching runnerID
	subject, err := internal.SubjectFromContext(ctx)
	if err != nil {
		return err
	}
	var isAgent bool
	switch s := subject.(type) {
	case *manager:
		// ok
	case *RunnerMeta:
		if s.ID != runnerID {
			return internal.ErrAccessNotPermitted
		}
		isAgent = true
	default:
		return internal.ErrAccessNotPermitted
	}

	// keep a record of what the status was before the update for logging
	// purposes
	var from RunnerStatus
	err = s.db.update(ctx, runnerID, func(runner *RunnerMeta) error {
		from = runner.Status
		return runner.setStatus(to, isAgent)
	})
	if err != nil {
		s.Error(err, "updating runner status", "runner_id", runnerID, "status", to, "subject", subject)
		return err
	}
	if isAgent {
		s.V(9).Info("updated runner status", "runner_id", runnerID, "from", from, "to", to, "subject", subject)
	}
	return nil
}

func (s *Service) listRunners(ctx context.Context) ([]*RunnerMeta, error) {
	return s.db.list(ctx)
}

func (s *Service) listServerRunners(ctx context.Context) ([]*RunnerMeta, error) {
	return s.db.listServerRunners(ctx)
}

func (s *Service) listRunnersByOrganization(ctx context.Context, organization string) ([]*RunnerMeta, error) {
	_, err := s.organization.CanAccess(ctx, rbac.ListRunnersAction, organization)
	if err != nil {
		return nil, err
	}
	return s.db.listRunnersByOrganization(ctx, organization)
}

func (s *Service) listRunnersByPool(ctx context.Context, poolID string) ([]*RunnerMeta, error) {
	return s.db.listRunnersByPool(ctx, poolID)
}

func (s *Service) deleteRunner(ctx context.Context, runnerID string) error {
	if err := s.db.deleteRunner(ctx, runnerID); err != nil {
		s.Error(err, "deleting runner", "runner_id", runnerID)
		return err
	}
	s.V(2).Info("deleted runner", "runner_id", runnerID)
	return nil
}

func (s *Service) createJob(ctx context.Context, run *otfrun.Run) error {
	job := newJob(run)
	if err := s.db.createJob(ctx, job); err != nil {
		return err
	}
	return nil
}

// cancelJob is called when a user cancels a run - cancelJob determines whether
// the corresponding job is signaled and what type of signal, and/or whether the
// job should be canceled.
func (s *Service) cancelJob(ctx context.Context, run *otfrun.Run) error {
	var (
		spec   = JobSpec{RunID: run.ID, Phase: run.Phase()}
		signal *bool
	)
	job, err := s.db.updateJob(ctx, spec, func(job *Job) (err error) {
		signal, err = job.cancel(run)
		return err
	})
	if err != nil {
		if errors.Is(err, internal.ErrResourceNotFound) {
			// ignore when no job has yet been created for the run.
			return nil
		}
		s.Error(err, "canceling job", "spec", spec)
		return err
	}
	if signal != nil {
		s.V(4).Info("sending cancelation signal to job", "force-cancel", *signal, "job", job)
	} else {
		s.V(4).Info("canceled job", "job", job)
	}
	return nil
}

// getJobs returns jobs that either:
// (a) have JobAllocated status
// (b) have JobRunning status and a non-nil cancellation signal
//
// getJobs is intended to be called by an runner in order to retrieve jobs to
// execute and jobs to cancel.
func (s *Service) getJobs(ctx context.Context, runnerID string) ([]*Job, error) {
	// only these subjects may call this endpoint:
	// (a) a runner with an ID matching runnerID
	if err := authorizeRunner(ctx, runnerID); err != nil {
		return nil, internal.ErrAccessNotPermitted
	}

	sub, unsub := s.WatchJobs(ctx)
	defer unsub()

	jobs, err := s.db.getAllocatedAndSignaledJobs(ctx, runnerID)
	if err != nil {
		return nil, err
	}

	if len(jobs) > 0 {
		// return existing jobs
		return jobs, nil
	}

	// wait for a job matching criteria to arrive:
	for event := range sub {
		job := event.Payload
		if job.RunnerID == nil || *job.RunnerID != runnerID {
			continue
		}
		switch job.Status {
		case JobAllocated:
			return []*Job{job}, nil
		case JobRunning:
			if job.Signaled != nil {
				return []*Job{job}, nil
			}
		}
	}
	return nil, nil
}

func (s *Service) getJob(ctx context.Context, spec JobSpec) (*Job, error) {
	return s.db.getJob(ctx, spec)
}

func (s *Service) listJobs(ctx context.Context) ([]*Job, error) {
	return s.db.listJobs(ctx)
}

func (s *Service) allocateJob(ctx context.Context, spec JobSpec, runnerID string) (*Job, error) {
	allocated, err := s.db.updateJob(ctx, spec, func(job *Job) error {
		return job.allocate(runnerID)
	})
	if err != nil {
		s.Error(err, "allocating job", "spec", spec, "runner_id", runnerID)
		return nil, err
	}
	s.V(0).Info("allocated job", "job", allocated, "runner_id", runnerID)
	return allocated, nil
}

func (s *Service) reallocateJob(ctx context.Context, spec JobSpec, runnerID string) (*Job, error) {
	var (
		from        string // ID of runner that job *was* allocated to
		reallocated *Job
	)
	reallocated, err := s.db.updateJob(ctx, spec, func(job *Job) error {
		from = *job.RunnerID
		return job.reallocate(runnerID)
	})
	if err != nil {
		s.Error(err, "re-allocating job", "spec", spec, "from", from, "to", runnerID)
		return nil, err
	}
	s.V(0).Info("re-allocated job", "spec", spec, "from", from, "to", runnerID)
	return reallocated, nil
}

// startJob starts a job and returns a job token with permissions to
// carry out the job. Only a runner that has been allocated the job can
// call this method.
func (s *Service) startJob(ctx context.Context, spec JobSpec) ([]byte, error) {
	subject, err := runnerFromContext(ctx)
	if err != nil {
		return nil, internal.ErrAccessNotPermitted
	}

	var token []byte
	_, err = s.db.updateJob(ctx, spec, func(job *Job) error {
		if job.RunnerID == nil || *job.RunnerID != subject.String() {
			return internal.ErrAccessNotPermitted
		}
		if err := job.startJob(); err != nil {
			return err
		}
		// start corresponding run phase too
		if _, err = s.phases.StartPhase(ctx, spec.RunID, spec.Phase, otfrun.PhaseStartOptions{}); err != nil {
			return err
		}
		token, err = s.createJobToken(spec)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		s.Error(err, "starting job", "spec", spec, "runner", subject)
		return nil, err
	}
	s.V(2).Info("started job", "spec", spec, "runner", subject)
	return token, nil
}

type finishJobOptions struct {
	Status JobStatus `json:"status"`
	Error  string    `json:"error,omitempty"`
}

// finishJob finishes a job. Only the job itself may call this endpoint.
func (s *Service) finishJob(ctx context.Context, spec JobSpec, opts finishJobOptions) error {
	{
		subject, err := internal.SubjectFromContext(ctx)
		if err != nil {
			return internal.ErrAccessNotPermitted
		}
		if _, ok := subject.(*Job); !ok {
			return internal.ErrAccessNotPermitted
		}
	}
	job, err := s.db.updateJob(ctx, spec, func(job *Job) error {
		// update corresponding run phase too
		var err error
		switch opts.Status {
		case JobFinished, JobErrored:
			_, err = s.phases.FinishPhase(ctx, spec.RunID, spec.Phase, otfrun.PhaseFinishOptions{
				Errored: opts.Status == JobErrored,
			})
		case JobCanceled:
			err = s.phases.Cancel(ctx, spec.RunID)
		}
		if err != nil {
			return err
		}
		return job.finishJob(opts.Status)
	})
	if err != nil {
		s.Error(err, "finishing job", "spec", spec)
		return err
	}
	if opts.Error != "" {
		s.V(2).Info("finished job with error", "job", job, "status", opts.Status, "job_error", opts.Error)
	} else {
		s.V(2).Info("finished job", "job", job, "status", opts.Status)
	}
	return nil
}

// agent tokens

func (s *Service) CreateAgentToken(ctx context.Context, poolID string, opts CreateAgentTokenOptions) (*agentToken, []byte, error) {
	at, token, subject, err := func() (*agentToken, []byte, internal.Subject, error) {
		pool, err := s.db.getPool(ctx, poolID)
		if err != nil {
			return nil, nil, nil, err
		}
		subject, err := s.organization.CanAccess(ctx, rbac.CreateAgentTokenAction, pool.Organization)
		if err != nil {
			return nil, nil, nil, err
		}
		at, token, err := s.NewAgentToken(poolID, opts)
		if err != nil {
			return nil, nil, nil, err
		}
		if err := s.db.createAgentToken(ctx, at); err != nil {
			s.Error(err, "creating agent token", "organization", poolID, "id", at.ID, "subject", subject)
			return nil, nil, nil, err
		}
		return at, token, subject, nil
	}()
	if err != nil {
		s.Error(err, "creating agent token", "agent_pool_id", poolID, "subject", subject)
		return nil, nil, err
	}
	s.V(0).Info("created agent token", "token", at, "subject", subject)
	return at, token, nil
}

func (s *Service) GetAgentToken(ctx context.Context, tokenID string) (*agentToken, error) {
	at, subject, err := func() (*agentToken, internal.Subject, error) {
		at, err := s.db.getAgentTokenByID(ctx, tokenID)
		if err != nil {
			return nil, nil, err
		}
		pool, err := s.db.getPool(ctx, at.AgentPoolID)
		if err != nil {
			return nil, nil, err
		}
		subject, err := s.organization.CanAccess(ctx, rbac.GetAgentTokenAction, pool.Organization)
		if err != nil {
			return nil, nil, err
		}
		return at, subject, nil
	}()
	if err != nil {
		s.Error(err, "retrieving agent token", "id", tokenID)
		return nil, err
	}
	s.V(9).Info("retrieved agent token", "token", at, "subject", subject)
	return at, nil
}

func (s *Service) ListAgentTokens(ctx context.Context, poolID string) ([]*agentToken, error) {
	pool, err := s.db.getPool(ctx, poolID)
	if err != nil {
		return nil, err
	}
	subject, err := s.organization.CanAccess(ctx, rbac.ListAgentTokensAction, pool.Organization)
	if err != nil {
		return nil, err
	}

	tokens, err := s.db.listAgentTokens(ctx, poolID)
	if err != nil {
		s.Error(err, "listing agent tokens", "organization", poolID, "subject", subject)
		return nil, err
	}
	s.V(9).Info("listed agent tokens", "organization", poolID, "subject", subject)
	return tokens, nil
}

func (s *Service) DeleteAgentToken(ctx context.Context, tokenID string) (*agentToken, error) {
	at, subject, err := func() (*agentToken, internal.Subject, error) {
		// retrieve agent token and pool in order to get organization for authorization
		at, err := s.db.getAgentTokenByID(ctx, tokenID)
		if err != nil {
			return nil, nil, err
		}
		pool, err := s.db.getPool(ctx, at.AgentPoolID)
		if err != nil {
			return nil, nil, err
		}
		subject, err := s.organization.CanAccess(ctx, rbac.DeleteAgentTokenAction, pool.Organization)
		if err != nil {
			return nil, nil, err
		}
		if err := s.db.deleteAgentToken(ctx, tokenID); err != nil {
			return nil, subject, err
		}
		return at, subject, nil
	}()
	if err != nil {
		s.Error(err, "deleting agent token", "id", tokenID)
		return nil, err
	}

	s.V(0).Info("deleted agent token", "token", at, "subject", subject)
	return at, nil
}

// pools

// checkWorkspacePoolAccess checks if a workspace has been granted access to a pool. If the
// pool is organization-scoped then the workspace automatically has access;
// otherwise access must already have been granted explicity.
func (s *Service) checkWorkspacePoolAccess(ctx context.Context, ws *workspace.Workspace) error {
	if ws.AgentPoolID == nil {
		// workspace is not using any pool
		return nil
	}
	pool, err := s.GetAgentPool(ctx, *ws.AgentPoolID)
	if err != nil {
		return err
	}
	if pool.OrganizationScoped {
		return nil
	} else if slices.Contains(pool.AllowedWorkspaces, ws.ID) {
		// is explicitly granted
		return nil
	}
	return ErrWorkspaceNotAllowedToUsePool
}

func (s *Service) CreateAgentPool(ctx context.Context, opts CreateAgentPoolOptions) (*Pool, error) {
	subject, err := s.organization.CanAccess(ctx, rbac.CreateAgentPoolAction, opts.Organization)
	if err != nil {
		return nil, err
	}
	pool, err := func() (*Pool, error) {
		pool, err := newPool(opts)
		if err != nil {
			return nil, err
		}
		if err := s.db.createPool(ctx, pool); err != nil {
			return nil, err
		}
		return pool, nil
	}()
	if err != nil {
		s.Error(err, "creating agent pool", "subject", subject)
		return nil, err
	}
	s.V(0).Info("created agent pool", "subject", subject, "pool", pool)
	return pool, nil
}

func (s *Service) updateAgentPool(ctx context.Context, poolID string, opts updatePoolOptions) (*Pool, error) {
	var (
		subject       internal.Subject
		before, after Pool
	)
	err := s.db.Lock(ctx, "agent_pools, agent_pool_allowed_workspaces", func(ctx context.Context, q *sqlc.Queries) (err error) {
		pool, err := s.db.getPool(ctx, poolID)
		if err != nil {
			return err
		}
		subject, err = s.organization.CanAccess(ctx, rbac.UpdateAgentPoolAction, pool.Organization)
		if err != nil {
			return err
		}
		before = *pool
		after = *pool
		if err := after.update(opts); err != nil {
			return err
		}
		if err := s.db.updatePool(ctx, &after); err != nil {
			return err
		}
		// Add/remove allowed workspaces
		add := internal.DiffStrings(after.AllowedWorkspaces, before.AllowedWorkspaces)
		remove := internal.DiffStrings(before.AllowedWorkspaces, after.AllowedWorkspaces)
		for _, workspaceID := range add {
			if err := s.db.addAgentPoolAllowedWorkspace(ctx, poolID, workspaceID); err != nil {
				return err
			}
		}
		for _, workspaceID := range remove {
			if err := s.db.deleteAgentPoolAllowedWorkspace(ctx, poolID, workspaceID); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		s.Error(err, "updating agent pool", "agent_pool_id", poolID, "subject", subject)
		return nil, err
	}
	s.V(0).Info("updated agent pool", "subject", subject, "before", &before, "after", &after)
	return &after, nil
}

func (s *Service) GetAgentPool(ctx context.Context, poolID string) (*Pool, error) {
	pool, err := s.db.getPool(ctx, poolID)
	if err != nil {
		s.Error(err, "retrieving agent pool", "agent_pool_id", poolID)
		return nil, err
	}
	subject, err := s.organization.CanAccess(ctx, rbac.GetAgentPoolAction, pool.Organization)
	if err != nil {
		return nil, err
	}
	s.V(9).Info("retrieved agent pool", "subject", subject, "organization", pool.Organization)
	return pool, nil
}

func (s *Service) listAgentPoolsByOrganization(ctx context.Context, organization string, opts listPoolOptions) ([]*Pool, error) {
	subject, err := s.organization.CanAccess(ctx, rbac.ListAgentPoolsAction, organization)
	if err != nil {
		return nil, err
	}
	pools, err := s.db.listPoolsByOrganization(ctx, organization, opts)
	if err != nil {
		s.Error(err, "listing agent pools", "subject", subject)
		return nil, err
	}
	s.V(9).Info("listed agent pools", "subject", subject, "count", len(pools))
	return pools, nil
}

func (s *Service) deleteAgentPool(ctx context.Context, poolID string) (*Pool, error) {
	pool, subject, err := func() (*Pool, internal.Subject, error) {
		// retrieve pool in order to get organization for authorization
		pool, err := s.db.getPool(ctx, poolID)
		if err != nil {
			return nil, nil, err
		}
		subject, err := s.organization.CanAccess(ctx, rbac.DeleteAgentPoolAction, pool.Organization)
		if err != nil {
			return nil, nil, err
		}
		// only permit pool to be deleted if it is not referenced by any
		// workspaces (it would raise a foreign key error anyway but friendlier
		// to return an informative error message).
		if len(pool.AssignedWorkspaces) > 0 {
			return nil, nil, ErrCannotDeletePoolReferencedByWorkspaces
		}
		if err := s.db.deleteAgentPool(ctx, pool.ID); err != nil {
			return nil, subject, err
		}
		return pool, subject, nil
	}()
	if err != nil {
		s.Error(err, "deleting agent pool", "agent_pool_id", poolID, "subject", subject)
		return nil, err
	}
	s.V(9).Info("deleted agent pool", "pool", pool, "subject", subject)
	return pool, nil
}
