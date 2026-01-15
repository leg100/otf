package runner

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/dynamiccreds"
	otfhttp "github.com/leg100/otf/internal/http"
	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/pubsub"
	"github.com/leg100/otf/internal/resource"
	otfrun "github.com/leg100/otf/internal/run"
	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/tfeapi"
	"github.com/leg100/otf/internal/tokens"
	"github.com/leg100/otf/internal/workspace"
)

var ErrInvalidStateTransition = errors.New("invalid runner state transition")

type (
	Service struct {
		logr.Logger
		*authz.Authorizer

		tfeapi       *tfe
		api          *api
		poolBroker   pubsub.SubscriptionService[*Pool]
		runnerBroker pubsub.SubscriptionService[*RunnerEvent]
		jobBroker    pubsub.SubscriptionService[*JobEvent]
		phases       phaseClient
		Signaler     *jobSignaler
		workspaces   *workspace.Service
		dynamiccreds *dynamiccreds.Service
		hostnames    *internal.HostnameService

		db *db
		*tokenFactory
	}

	ServiceOptions struct {
		logr.Logger
		*sql.DB
		*sql.Listener
		*tfeapi.Responder

		RunService                *otfrun.Service
		WorkspaceService          *workspace.Service
		TokensService             *tokens.Service
		Authorizer                *authz.Authorizer
		DynamicCredentialsService *dynamiccreds.Service
		HostnameService           *internal.HostnameService
	}

	phaseClient interface {
		StartPhase(ctx context.Context, runID resource.TfeID, phase otfrun.PhaseType, _ otfrun.PhaseStartOptions) (*otfrun.Run, error)
		FinishPhase(ctx context.Context, runID resource.TfeID, phase otfrun.PhaseType, opts otfrun.PhaseFinishOptions) (*otfrun.Run, error)
		Cancel(ctx context.Context, runID resource.TfeID) error
	}
)

func NewService(opts ServiceOptions) *Service {
	svc := &Service{
		Logger:       opts.Logger,
		Authorizer:   opts.Authorizer,
		db:           &db{DB: opts.DB},
		phases:       opts.RunService,
		Signaler:     newJobSignaler(opts.Logger, opts.DB),
		workspaces:   opts.WorkspaceService,
		dynamiccreds: opts.DynamicCredentialsService,
		hostnames:    opts.HostnameService,
	}
	svc.tokenFactory = &tokenFactory{
		tokens: opts.TokensService,
	}
	svc.tfeapi = &tfe{
		Service:   svc,
		Responder: opts.Responder,
	}
	svc.api = &api{
		Service:   svc,
		Responder: opts.Responder,
	}
	svc.poolBroker = pubsub.NewBroker[*Pool](
		opts.Logger,
		opts.Listener,
		"agent_pools",
	)
	svc.runnerBroker = pubsub.NewBroker[*RunnerEvent](
		opts.Logger,
		opts.Listener,
		"runners",
	)
	svc.jobBroker = pubsub.NewBroker[*JobEvent](
		opts.Logger,
		opts.Listener,
		"jobs",
	)
	// Register with auth middleware the agent token kind and a means of
	// retrieving the appropriate runner corresponding to the agent token ID
	opts.TokensService.RegisterKind(resource.AgentTokenKind, func(ctx context.Context, tokenID resource.TfeID) (authz.Subject, error) {
		// Fetch agent pool corresponding to the provided token. This
		// effectively authenticates the token.
		pool, err := svc.db.getPoolByTokenID(ctx, tokenID)
		if err != nil {
			return nil, err
		}
		// if the runner has already registered then it should be sending its ID
		// in an http header
		headers, err := otfhttp.HeadersFromContext(ctx)
		if err != nil {
			return nil, err
		}
		if runnerIDValue := headers.Get(runnerIDHeaderKey); runnerIDValue != "" {
			runnerID, err := resource.ParseTfeID(runnerIDValue)
			if err != nil {
				return nil, err
			}
			runner, err := svc.getRunner(ctx, runnerID)
			if err != nil {
				return nil, fmt.Errorf("retrieving runner corresponding to ID found in http header: %w", err)
			}
			return runner, nil
		}
		// Agent runner hasn't registered yet, so set subject to a runner with a
		// agent pool info, which will be used when registering the runner below.
		return &unregistered{pool: pool}, nil
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
	opts.TokensService.RegisterKind(resource.JobKind, func(ctx context.Context, jobID resource.TfeID) (authz.Subject, error) {
		return svc.GetJob(ctx, jobID)
	})
	return svc
}

func (s *Service) AddHandlers(r *mux.Router) {
	s.tfeapi.addHandlers(r)
	s.api.addHandlers(r)
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

func (s *Service) WatchRunners(ctx context.Context) (<-chan pubsub.Event[*RunnerEvent], func()) {
	return s.runnerBroker.Subscribe(ctx)
}

func (s *Service) Watch(ctx context.Context) (<-chan pubsub.Event[*RunnerEvent], func()) {
	return s.WatchRunners(ctx)
}

func (s *Service) WatchJobs(ctx context.Context) (<-chan pubsub.Event[*JobEvent], func()) {
	return s.jobBroker.Subscribe(ctx)
}

func (s *Service) Register(ctx context.Context, opts RegisterRunnerOptions) (*RunnerMeta, error) {
	runner, err := func() (*RunnerMeta, error) {
		subject, err := authz.SubjectFromContext(ctx)
		if err != nil {
			return nil, err
		}
		unregistered, ok := subject.(*unregistered)
		if !ok {
			return nil, internal.ErrAccessNotPermitted
		}
		registered, err := register(unregistered, opts)
		if err != nil {
			return nil, err
		}
		if err := s.db.create(ctx, registered); err != nil {
			return nil, err
		}
		return registered, nil
	}()
	if err != nil {
		s.Error(err, "registering runner")
		return nil, err
	}
	s.V(0).Info("registered runner", "runner", runner)
	return runner, nil
}

func (s *Service) getRunner(ctx context.Context, runnerID resource.TfeID) (*RunnerMeta, error) {
	runner, err := s.db.get(ctx, runnerID)
	if err != nil {
		s.Error(err, "retrieving runner", "runner_id", runnerID)
		return nil, err
	}
	s.V(9).Info("retrieved runner", "runner", runner)
	return runner, err
}

func (s *Service) updateStatus(ctx context.Context, runnerID resource.TfeID, to RunnerStatus) error {
	// only these subjects may call this endpoint:
	// (a) the manager, or
	// (b) an runner with an ID matching runnerID
	subject, err := authz.SubjectFromContext(ctx)
	if err != nil {
		return err
	}
	var ping bool
	switch s := subject.(type) {
	case *manager:
		// ok
	case *RunnerMeta:
		if s.ID != runnerID {
			return internal.ErrAccessNotPermitted
		}
		ping = true
	default:
		return internal.ErrAccessNotPermitted
	}

	// keep a record of what the status was before the update for logging
	// purposes
	var from RunnerStatus
	err = s.db.update(ctx, runnerID, func(ctx context.Context, runner *RunnerMeta) error {
		from = runner.Status
		return runner.setStatus(to, ping)
	})
	if err != nil {
		s.Error(err, "updating runner status", "runner_id", runnerID, "status", to, "subject", subject)
		return err
	}
	s.V(9).Info("updated runner status", "runner_id", runnerID, "from", from, "to", to, "subject", subject)
	return nil
}

type ListOptions struct {
	resource.PageOptions
	// Organization filters runners by the organization of their agent pool.
	//
	// NOTE: setting this does not exclude server runners (which do not belong
	// to an organization). To exclude servers runners as well set Server below to
	// false.
	Organization *organization.Name `schema:"organization_name"`
	// PoolID filters runners by agent pool ID
	PoolID *resource.TfeID `schema:"pool_id"`
	// HideServerRunners if true filters out server runners
	HideServerRunners bool `schema:"hide_server_runners"`
}

func (s *Service) List(ctx context.Context, opts ListOptions) (*resource.Page[*RunnerMeta], error) {
	runners, err := s.ListRunners(ctx, opts)
	if err != nil {
		return nil, err
	}
	return resource.NewPage(runners, opts.PageOptions, nil), nil
}

func (s *Service) ListRunners(ctx context.Context, opts ListOptions) ([]*RunnerMeta, error) {
	// Set scope in which caller is requesting to list runners.
	var scope resource.ID
	if opts.PoolID != nil {
		scope = opts.PoolID
	} else if opts.Organization != nil {
		scope = opts.Organization
	} else {
		scope = resource.SiteID
	}
	// Check if caller has perms to list runners in that scope.
	if _, err := s.Authorize(ctx, authz.ListRunnersAction, scope); err != nil {
		return nil, err
	}
	return s.db.list(ctx, opts)
}

func (s *Service) DeleteRunner(ctx context.Context, runnerID resource.TfeID) error {
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
		signal bool
		force  bool
		err    error
	)
	job, err := s.db.updateJob(ctx, run.ID, func(ctx context.Context, job *Job) error {
		// Determine whether job is to be signaled and if so if it is to be
		// signaled by force.
		signal, force, err = job.cancel(run)
		if err != nil {
			return err
		}
		if signal {
			if err := s.Signaler.publish(ctx, job.ID, force); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		if errors.Is(err, internal.ErrResourceNotFound) {
			// Don't raise error when no job has yet been created for the run.
			return nil
		}
		s.Error(err, "canceling job for run", "run", run)
		return err
	}
	if signal {
		s.V(4).Info("sent cancelation signal to job", "force", force, "job", job)
	}
	if job.Status == JobCanceled {
		s.V(4).Info("canceled job", "job", job)
	}
	return nil
}

// awaitAllocatedJobs waits until there is at least one allocated job for a
// runner before returning allocated job(s).
func (s *Service) awaitAllocatedJobs(ctx context.Context, runnerID resource.TfeID) ([]*Job, error) {
	// only these subjects may call this endpoint:
	// (a) a runner with an ID matching runnerID
	if err := authorizeRunner(ctx, runnerID); err != nil {
		return nil, internal.ErrAccessNotPermitted
	}

	sub, unsub := s.WatchJobs(ctx)
	defer unsub()

	jobs, err := s.db.listAllocatedJobs(ctx, runnerID)
	if err != nil {
		return nil, err
	}

	if len(jobs) > 0 {
		// return existing jobs
		return jobs, nil
	}

	// wait for a job matching criteria to arrive:
	for event := range sub {
		if event.Payload.RunnerID == nil || *event.Payload.RunnerID != runnerID {
			continue
		}
		var match bool
		switch event.Payload.Status {
		case JobAllocated:
			match = true
		case JobRunning:
			if event.Payload.Signaled != nil {
				match = true
			}
		}
		if match {
			job, err := s.GetJob(ctx, event.Payload.ID)
			if err != nil {
				return nil, fmt.Errorf("retrieving job: %w", err)
			}
			return []*Job{job}, nil
		}
	}
	return nil, nil
}

func (s *Service) GetJob(ctx context.Context, jobID resource.TfeID) (*Job, error) {
	return s.db.getJob(ctx, jobID)
}

func (s *Service) awaitJobSignal(ctx context.Context, jobID resource.TfeID) func() (jobSignal, error) {
	return s.Signaler.awaitJobSignal(ctx, jobID)
}

func (s *Service) listJobs(ctx context.Context) ([]*Job, error) {
	return s.db.listJobs(ctx)
}

func (s *Service) allocateJob(ctx context.Context, jobID resource.TfeID, runnerID resource.TfeID) (*Job, error) {
	allocated, err := s.db.updateJob(ctx, jobID, func(ctx context.Context, job *Job) error {
		return job.allocate(runnerID)
	})
	if err != nil {
		s.Error(err, "allocating job", "job_id", jobID, "runner_id", runnerID)
		return nil, err
	}
	s.V(0).Info("allocated job", "job", allocated, "runner_id", runnerID)
	return allocated, nil
}

func (s *Service) reallocateJob(ctx context.Context, jobID resource.TfeID, runnerID resource.TfeID) (*Job, error) {
	var (
		from        resource.TfeID // ID of runner that job *was* allocated to
		reallocated *Job
	)
	reallocated, err := s.db.updateJob(ctx, jobID, func(ctx context.Context, job *Job) error {
		from = *job.RunnerID
		return job.reallocate(runnerID)
	})
	if err != nil {
		s.Error(err, "re-allocating job", "job_id", jobID, "from", from, "to", runnerID)
		return nil, err
	}
	s.V(0).Info("re-allocated job", "job", reallocated, "from", from, "to", runnerID)
	return reallocated, nil
}

// startJob starts a job and returns a job token with permissions to
// carry out the job. Only a runner that has been allocated the job can
// call this method.
func (s *Service) startJob(ctx context.Context, jobID resource.TfeID) ([]byte, error) {
	runner, err := runnerFromContext(ctx)
	if err != nil {
		return nil, internal.ErrAccessNotPermitted
	}

	var token []byte
	_, err = s.db.updateJob(ctx, jobID, func(ctx context.Context, job *Job) error {
		if job.RunnerID == nil || *job.RunnerID != runner.ID {
			return internal.ErrAccessNotPermitted
		}
		if err := job.startJob(); err != nil {
			return err
		}
		// start corresponding run phase too
		if _, err = s.phases.StartPhase(ctx, job.RunID, job.Phase, otfrun.PhaseStartOptions{}); err != nil {
			return err
		}
		token, err = s.createJobToken(jobID)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		s.Error(err, "starting job", "spec", jobID, "runner", runner)
		return nil, err
	}
	s.V(2).Info("started job", "spec", jobID, "runner", runner)
	return token, nil
}

type finishJobOptions struct {
	Status JobStatus `json:"status"`
	Error  string    `json:"error,omitempty"`
}

// finishJob finishes a job. Only the job itself may call this endpoint.
func (s *Service) finishJob(ctx context.Context, jobID resource.TfeID, opts finishJobOptions) error {
	{
		subject, err := authz.SubjectFromContext(ctx)
		if err != nil {
			return internal.ErrAccessNotPermitted
		}
		if _, ok := subject.(*Job); !ok {
			return internal.ErrAccessNotPermitted
		}
	}
	job, err := s.db.updateJob(ctx, jobID, func(ctx context.Context, job *Job) error {
		// update corresponding run phase too
		var err error
		switch opts.Status {
		case JobFinished, JobErrored:
			_, err = s.phases.FinishPhase(ctx, job.RunID, job.Phase, otfrun.PhaseFinishOptions{
				Errored: opts.Status == JobErrored,
			})
		case JobCanceled:
			err = s.phases.Cancel(ctx, job.RunID)
		}
		if err != nil {
			return err
		}
		return job.finishJob(opts.Status)
	})
	if err != nil {
		s.Error(err, "finishing job", "job_id", jobID)
		return err
	}
	if opts.Error != "" {
		s.V(2).Info("finished job with error", "job", job, "status", opts.Status, "job_error", opts.Error)
	} else {
		s.V(2).Info("finished job", "job", job, "status", opts.Status)
	}
	return nil
}

func (s *Service) GenerateDynamicCredentialsToken(ctx context.Context, jobID resource.TfeID, audience string) ([]byte, error) {
	token, err := func() ([]byte, error) {
		if s.dynamiccreds.PrivateKey() == nil {
			return nil, errors.New("no private key has been configured")
		}
		job, err := s.GetJob(ctx, jobID)
		if err != nil {
			return nil, err
		}
		workspace, err := s.workspaces.Get(ctx, job.WorkspaceID)
		if err != nil {
			return nil, err
		}
		return s.dynamiccreds.GenerateToken(
			s.hostnames.URL(""),
			job.Organization,
			job.WorkspaceID,
			workspace.Name,
			job.RunID,
			job.Phase,
			audience,
		)
	}()
	if err != nil {
		s.Error(err, "generating token for dynamic credentials", "job_id", jobID, "audience", audience)
		return nil, err
	}
	s.V(4).Info("generated token for dynamic credentials", "job_id", jobID, "audience", audience)
	return token, nil
}

// agent tokens

func (s *Service) CreateAgentToken(ctx context.Context, poolID resource.TfeID, opts CreateAgentTokenOptions) (*AgentToken, []byte, error) {
	at, token, subject, err := func() (*AgentToken, []byte, authz.Subject, error) {
		pool, err := s.db.getPool(ctx, poolID)
		if err != nil {
			return nil, nil, nil, err
		}
		subject, err := s.Authorize(ctx, authz.CreateAgentTokenAction, &pool.Organization)
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

func (s *Service) GetAgentToken(ctx context.Context, tokenID resource.TfeID) (*AgentToken, error) {
	at, subject, err := func() (*AgentToken, authz.Subject, error) {
		at, err := s.db.getAgentTokenByID(ctx, tokenID)
		if err != nil {
			return nil, nil, err
		}
		pool, err := s.db.getPool(ctx, at.AgentPoolID)
		if err != nil {
			return nil, nil, err
		}
		subject, err := s.Authorize(ctx, authz.GetAgentTokenAction, &pool.Organization)
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

func (s *Service) ListAgentTokens(ctx context.Context, poolID resource.TfeID) ([]*AgentToken, error) {
	pool, err := s.db.getPool(ctx, poolID)
	if err != nil {
		return nil, err
	}
	subject, err := s.Authorize(ctx, authz.ListAgentTokensAction, &pool.Organization)
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

func (s *Service) DeleteAgentToken(ctx context.Context, tokenID resource.TfeID) (*AgentToken, error) {
	at, subject, err := func() (*AgentToken, authz.Subject, error) {
		// retrieve agent token and pool in order to get organization for authorization
		at, err := s.db.getAgentTokenByID(ctx, tokenID)
		if err != nil {
			return nil, nil, err
		}
		pool, err := s.db.getPool(ctx, at.AgentPoolID)
		if err != nil {
			return nil, nil, err
		}
		subject, err := s.Authorize(ctx, authz.DeleteAgentTokenAction, &pool.Organization)
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
	subject, err := s.Authorize(ctx, authz.CreateAgentPoolAction, &opts.Organization)
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

func (s *Service) UpdateAgentPool(ctx context.Context, poolID resource.TfeID, opts UpdatePoolOptions) (*Pool, error) {
	var (
		subject       authz.Subject
		before, after Pool
	)
	err := s.db.Lock(ctx, "agent_pools, agent_pool_allowed_workspaces", func(ctx context.Context) (err error) {
		pool, err := s.db.getPool(ctx, poolID)
		if err != nil {
			return err
		}
		subject, err = s.Authorize(ctx, authz.UpdateAgentPoolAction, &pool.Organization)
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
		add := internal.Diff(after.AllowedWorkspaces, before.AllowedWorkspaces)
		remove := internal.Diff(before.AllowedWorkspaces, after.AllowedWorkspaces)
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

func (s *Service) GetAgentPool(ctx context.Context, poolID resource.TfeID) (*Pool, error) {
	pool, err := s.db.getPool(ctx, poolID)
	if err != nil {
		s.Error(err, "retrieving agent pool", "agent_pool_id", poolID)
		return nil, err
	}
	subject, err := s.Authorize(ctx, authz.GetAgentPoolAction, &pool.Organization)
	if err != nil {
		return nil, err
	}
	s.V(9).Info("retrieved agent pool", "subject", subject, "organization", pool.Organization)
	return pool, nil
}

func (s *Service) ListAgentPoolsByOrganization(ctx context.Context, organization organization.Name, opts ListPoolOptions) ([]*Pool, error) {
	subject, err := s.Authorize(ctx, authz.ListAgentPoolsAction, organization)
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

func (s *Service) DeleteAgentPool(ctx context.Context, poolID resource.TfeID) (*Pool, error) {
	pool, subject, err := func() (*Pool, authz.Subject, error) {
		// retrieve pool in order to get organization for authorization
		pool, err := s.db.getPool(ctx, poolID)
		if err != nil {
			return nil, nil, err
		}
		subject, err := s.Authorize(ctx, authz.DeleteAgentPoolAction, &pool.Organization)
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
