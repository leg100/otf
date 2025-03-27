package run

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/configversion"
	"github.com/leg100/otf/internal/logs"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/pubsub"
	"github.com/leg100/otf/internal/releases"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/runstatus"
	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/tfeapi"
	"github.com/leg100/otf/internal/tokens"
	"github.com/leg100/otf/internal/user"
	"github.com/leg100/otf/internal/vcs"
	"github.com/leg100/otf/internal/vcsprovider"
	"github.com/leg100/otf/internal/workspace"
	"github.com/leg100/surl/v2"
)

type (
	// Alias services so they don't conflict when nested together in struct
	ConfigurationVersionService configversion.Service
	VCSProviderService          vcsprovider.Service

	Service struct {
		logr.Logger
		authz.Interface

		workspaces             *workspace.Service
		cache                  internal.Cache
		db                     *pgdb
		tfeapi                 *tfe
		api                    *api
		web                    *webHandlers
		logs                   *logs.Service
		afterCancelHooks       []func(context.Context, *Run) error
		afterForceCancelHooks  []func(context.Context, *Run) error
		afterEnqueuePlanHooks  []func(context.Context, *Run) error
		afterEnqueueApplyHooks []func(context.Context, *Run) error
		broker                 pubsub.SubscriptionService[*Run]

		*factory
	}

	Options struct {
		Authorizer         *authz.Authorizer
		VCSEventSubscriber vcs.Subscriber

		WorkspaceService     *workspace.Service
		OrganizationService  *organization.Service
		ConfigVersionService *configversion.Service
		ReleasesService      *releases.Service
		VCSProviderService   *vcsprovider.Service
		TokensService        *tokens.Service
		LogsService          *logs.Service

		logr.Logger
		internal.Cache
		*sql.DB
		*tfeapi.Responder
		*surl.Signer
		*sql.Listener
	}
)

func NewService(opts Options) *Service {
	db := &pgdb{opts.DB}
	svc := Service{
		Logger:     opts.Logger,
		workspaces: opts.WorkspaceService,
		db:         db,
		cache:      opts.Cache,
		Interface:  opts.Authorizer,
		logs:       opts.LogsService,
	}
	svc.factory = &factory{
		organizations: opts.OrganizationService,
		workspaces:    opts.WorkspaceService,
		configs:       opts.ConfigVersionService,
		vcs:           opts.VCSProviderService,
		releases:      opts.ReleasesService,
	}
	svc.web = newWebHandlers(&svc, opts)
	svc.tfeapi = &tfe{
		Service:    &svc,
		workspaces: opts.WorkspaceService,
		Responder:  opts.Responder,
		Signer:     opts.Signer,
		authorizer: opts.Authorizer,
	}
	svc.api = &api{
		Service:   &svc,
		Responder: opts.Responder,
		Logger:    opts.Logger,
	}
	spawner := &Spawner{
		Logger:     opts.Logger.WithValues("component", "spawner"),
		configs:    opts.ConfigVersionService,
		workspaces: opts.WorkspaceService,
		vcs:        opts.VCSProviderService,
		runs:       &svc,
	}
	svc.broker = pubsub.NewBroker(
		opts.Logger,
		opts.Listener,
		"runs",
		func(ctx context.Context, id resource.ID, action sql.Action) (*Run, error) {
			if action == sql.DeleteAction {
				return &Run{ID: id}, nil
			}
			return db.GetRun(ctx, id)
		},
	)

	// Fetch related resources when API requests their inclusion
	opts.Responder.Register(tfeapi.IncludeCreatedBy, svc.tfeapi.includeCreatedBy)
	opts.Responder.Register(tfeapi.IncludeCurrentRun, svc.tfeapi.includeCurrentRun)

	// Resolve authorization requests for run IDs to a workspace IDs
	opts.Authorizer.RegisterWorkspaceResolver(resource.RunKind,
		func(ctx context.Context, runID resource.ID) (resource.ID, error) {
			run, err := db.GetRun(ctx, runID)
			if err != nil {
				return resource.ID{}, err
			}
			return run.WorkspaceID, nil
		},
	)

	// Subscribe run spawner to incoming vcs events
	opts.VCSEventSubscriber.Subscribe(spawner.handle)

	// After a workspace is created, if auto-queue-runs is set, then create a
	// run as well.
	opts.WorkspaceService.AfterCreateWorkspace(svc.autoQueueRun)

	return &svc
}

func (s *Service) AddHandlers(r *mux.Router) {
	s.web.addHandlers(r)
	s.tfeapi.addHandlers(r)
	s.api.addHandlers(r)
}

func (s *Service) Create(ctx context.Context, workspaceID resource.ID, opts CreateOptions) (*Run, error) {
	subject, err := s.Authorize(ctx, authz.CreateRunAction, &authz.AccessRequest{ID: &workspaceID})
	if err != nil {
		return nil, err
	}

	run, err := s.NewRun(ctx, workspaceID, opts)
	if err != nil {
		s.Error(err, "constructing new run", "subject", subject)
		return nil, err
	}

	if err = s.db.CreateRun(ctx, run); err != nil {
		s.Error(err, "creating run", "id", run.ID, "workspace_id", run.WorkspaceID, "subject", subject)
		return nil, err
	}
	s.V(1).Info("created run", "id", run.ID, "workspace_id", run.WorkspaceID, "subject", subject)

	return run, nil
}

// Get retrieves a run from the db.
func (s *Service) Get(ctx context.Context, runID resource.ID) (*Run, error) {
	subject, err := s.Authorize(ctx, authz.GetRunAction, &authz.AccessRequest{ID: &runID})
	if err != nil {
		return nil, err
	}

	run, err := s.db.GetRun(ctx, runID)
	if err != nil {
		s.Error(err, "retrieving run", "id", runID, "subject", subject)
		return nil, err
	}
	s.V(9).Info("retrieved run", "id", runID, "subject", subject)

	return run, nil
}

// List retrieves multiple runs. Use opts to filter and paginate the
// list.
func (s *Service) List(ctx context.Context, opts ListOptions) (*resource.Page[*Run], error) {
	var (
		subject authz.Subject
		authErr error
	)
	if opts.Organization != nil && opts.WorkspaceName != nil {
		workspace, err := s.workspaces.GetByName(ctx, *opts.Organization, *opts.WorkspaceName)
		if err != nil {
			return nil, err
		}
		// subject needs perms on workspace to list runs in workspace
		subject, authErr = s.Authorize(ctx, authz.GetWorkspaceAction, &authz.AccessRequest{ID: &workspace.ID})
	} else if opts.WorkspaceID != nil {
		// subject needs perms on workspace to list runs in workspace
		subject, authErr = s.Authorize(ctx, authz.GetWorkspaceAction, &authz.AccessRequest{ID: opts.WorkspaceID})
	} else if opts.Organization != nil {
		// subject needs perms on org to list runs in org
		subject, authErr = s.Authorize(ctx, authz.ListRunsAction, &authz.AccessRequest{Organization: opts.Organization})
	} else {
		// subject needs to be site admin to list runs across site
		subject, authErr = s.Authorize(ctx, authz.ListRunsAction, nil)
	}
	if authErr != nil {
		return nil, authErr
	}

	page, err := s.db.ListRuns(ctx, opts)
	if err != nil {
		s.Error(err, "listing runs", "subject", subject)
		return nil, err
	}

	s.V(9).Info("listed runs", "count", len(page.Items), "subject", subject)

	return page, nil
}

// EnqueuePlan enqueues a plan for the run.
func (s *Service) EnqueuePlan(ctx context.Context, runID resource.ID) (run *Run, err error) {
	err = s.db.Tx(ctx, func(ctx context.Context, _ sql.Connection) error {
		run, err = s.db.UpdateStatus(ctx, runID, func(ctx context.Context, run *Run) error {
			return run.EnqueuePlan()
		})
		if err != nil {
			return err
		}
		if !run.PlanOnly {
			_, err := s.workspaces.Lock(ctx, run.WorkspaceID, &run.ID)
			if err != nil {
				return err
			}
			_, err = s.workspaces.SetLatestRun(ctx, run.WorkspaceID, run.ID)
			if err != nil {
				return err
			}
		}
		for _, hook := range s.afterEnqueuePlanHooks {
			if err := hook(ctx, run); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		s.Error(err, "enqueuing plan", "id", runID)
		return nil, err
	}
	s.V(0).Info("enqueued plan", "id", runID)
	return run, err
}

func (s *Service) AfterEnqueuePlan(hook func(context.Context, *Run) error) {
	// add hook to list of hooks to be triggered after plan is enqueued
	s.afterEnqueuePlanHooks = append(s.afterEnqueuePlanHooks, hook)
}

func (s *Service) Delete(ctx context.Context, runID resource.ID) error {
	subject, err := s.Authorize(ctx, authz.DeleteRunAction, &authz.AccessRequest{ID: &runID})
	if err != nil {
		return err
	}

	if err := s.db.DeleteRun(ctx, runID); err != nil {
		s.Error(err, "deleting run", "id", runID, "subject", subject)
		return err
	}
	s.V(0).Info("deleted run", "id", runID, "subject", subject)
	return nil
}

// StartPhase starts a run phase.
func (s *Service) StartPhase(ctx context.Context, runID resource.ID, phase internal.PhaseType, _ PhaseStartOptions) (*Run, error) {
	run, err := s.db.UpdateStatus(ctx, runID, func(ctx context.Context, run *Run) error {
		return run.Start()
	})
	if err != nil {
		// only log error if not an phase already started error - this occurs when
		// multiple agents 'race' to start the phase and only one can do so,
		// whereas the other agents receive this error which is a legitimate
		// error condition and not something that should be reported to the
		// user.
		if !errors.Is(err, ErrPhaseAlreadyStarted) {
			s.Error(err, "starting "+string(phase), "id", runID)
		}
		return nil, err
	}
	s.V(0).Info("started "+string(phase), "id", runID)
	return run, nil
}

// FinishPhase finishes a phase. Creates a report of changes before updating the status of
// the run.
func (s *Service) FinishPhase(ctx context.Context, runID resource.ID, phase internal.PhaseType, opts PhaseFinishOptions) (*Run, error) {
	var resourceReport, outputReport Report
	if !opts.Errored {
		var err error
		resourceReport, outputReport, err = s.createReports(ctx, runID, phase)
		if err != nil {
			s.Error(err, "creating report", "id", runID, "phase", phase)
			opts.Errored = true
		}
	}
	var run *Run
	err := s.db.Tx(ctx, func(ctx context.Context, _ sql.Connection) (err error) {
		var autoapply bool
		run, err = s.db.UpdateStatus(ctx, runID, func(ctx context.Context, run *Run) (err error) {
			autoapply, err = run.Finish(phase, opts)
			return err
		})
		if err != nil {
			return err
		}
		if autoapply {
			return s.Apply(ctx, runID)
		}
		return nil
	})
	if err != nil {
		s.Error(err, "finishing "+string(phase), "id", runID, "subject")
		return nil, err
	}
	s.V(0).Info("finished "+string(phase), "id", runID, "resource_changes", resourceReport, "output_changes", outputReport, "run_status", run.Status)
	return run, nil
}

func (s *Service) Watch(ctx context.Context) (<-chan pubsub.Event[*Run], func()) {
	return s.broker.Subscribe(ctx)
}

// watchWithOptions provides authenticated access to a stream of run events,
// with the option to filter events.
func (s *Service) watchWithOptions(ctx context.Context, opts WatchOptions) (<-chan pubsub.Event[*Run], error) {
	var err error
	if opts.WorkspaceID != nil {
		// caller must have workspace-level read permissions
		_, err = s.Authorize(ctx, authz.WatchAction, &authz.AccessRequest{ID: opts.WorkspaceID})
	} else if opts.Organization != nil {
		// caller must have organization-level read permissions
		_, err = s.Authorize(ctx, authz.WatchAction, &authz.AccessRequest{Organization: opts.Organization})
	} else {
		// caller must have site-level read permissions
		_, err = s.Authorize(ctx, authz.WatchAction, nil)
	}
	if err != nil {
		return nil, err
	}

	sub, _ := s.broker.Subscribe(ctx)
	// relay is returned to the caller to which filtered run events are sent
	relay := make(chan pubsub.Event[*Run])
	go func() {
		// relay events
		for event := range sub {
			// apply workspace filter
			if opts.WorkspaceID != nil {
				if event.Payload.WorkspaceID != *opts.WorkspaceID {
					continue
				}
			}
			// apply organization filter
			if opts.Organization != nil {
				if event.Payload.Organization != *opts.Organization {
					continue
				}
			}
			relay <- event
		}
		close(relay)
	}()
	return relay, nil
}

// Apply enqueues an apply for the run.
func (s *Service) Apply(ctx context.Context, runID resource.ID) error {
	subject, err := s.Authorize(ctx, authz.ApplyRunAction, &authz.AccessRequest{ID: &runID})
	if err != nil {
		return err
	}
	return s.db.Tx(ctx, func(ctx context.Context, _ sql.Connection) error {
		run, err := s.db.UpdateStatus(ctx, runID, func(ctx context.Context, run *Run) error {
			return run.EnqueueApply()
		})
		if err != nil {
			s.Error(err, "enqueuing apply", "id", runID, "subject", subject)
			return err
		}

		s.V(0).Info("enqueued apply", "id", runID, "subject", subject)
		// invoke AfterEnqueueApply hooks
		for _, hook := range s.afterEnqueueApplyHooks {
			if err := hook(ctx, run); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *Service) AfterEnqueueApply(hook func(context.Context, *Run) error) {
	// add hook to list of hooks to be triggered after apply is enqueued
	s.afterEnqueueApplyHooks = append(s.afterEnqueueApplyHooks, hook)
}

// Discard discards the run.
func (s *Service) Discard(ctx context.Context, runID resource.ID) error {
	subject, err := s.Authorize(ctx, authz.DiscardRunAction, &authz.AccessRequest{ID: &runID})
	if err != nil {
		return err
	}

	_, err = s.db.UpdateStatus(ctx, runID, func(ctx context.Context, run *Run) error {
		return run.Discard()
	})
	if err != nil {
		s.Error(err, "discarding run", "id", runID, "subject", subject)
		return err
	}

	s.V(0).Info("discarded run", "id", runID, "subject", subject)

	return err
}

func (s *Service) Cancel(ctx context.Context, runID resource.ID) error {
	subject, err := s.Authorize(ctx, authz.CancelRunAction, &authz.AccessRequest{ID: &runID})
	if err != nil {
		return err
	}
	return s.db.Tx(ctx, func(ctx context.Context, _ sql.Connection) error {
		_, isUser := subject.(*user.User)

		run, err := s.db.UpdateStatus(ctx, runID, func(ctx context.Context, run *Run) (err error) {
			return run.Cancel(isUser, false)
		})
		if err != nil {
			s.Error(err, "canceling run", "id", runID, "subject", subject)
			return err
		}
		if run.CancelSignaledAt != nil && run.Status != runstatus.Canceled {
			s.V(0).Info("sent cancelation signal to run", "id", runID, "subject", subject)
		} else {
			s.V(0).Info("canceled run", "id", runID, "subject", subject)
		}
		// invoke AfterCancel hooks
		for _, hook := range s.afterCancelHooks {
			if err := hook(ctx, run); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *Service) AfterCancelRun(hook func(context.Context, *Run) error) {
	// add hook to list of hooks to be triggered after run is canceled
	s.afterCancelHooks = append(s.afterCancelHooks, hook)
}

// ForceCancel forcefully cancels a run.
func (s *Service) ForceCancel(ctx context.Context, runID resource.ID) error {
	subject, err := s.Authorize(ctx, authz.ForceCancelRunAction, &authz.AccessRequest{ID: &runID})
	if err != nil {
		return err
	}
	return s.db.Tx(ctx, func(ctx context.Context, _ sql.Connection) error {
		run, err := s.db.UpdateStatus(ctx, runID, func(ctx context.Context, run *Run) (err error) {
			return run.Cancel(true, true)
		})
		if err != nil {
			s.Error(err, "force canceling run", "id", runID, "subject", subject)
			return err
		}
		s.V(0).Info("force canceled run", "id", runID, "subject", subject)
		// invoke AfterForceCancelRun hooks
		for _, hook := range s.afterForceCancelHooks {
			if err := hook(ctx, run); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *Service) AfterForceCancelRun(hook func(context.Context, *Run) error) {
	// add hook to list of hooks to be triggered after run is force canceled
	s.afterForceCancelHooks = append(s.afterForceCancelHooks, hook)
}

func planFileCacheKey(f PlanFormat, id resource.ID) string {
	return fmt.Sprintf("%s.%s", id, f)
}

// GetPlanFile returns the plan file for the run.
func (s *Service) GetPlanFile(ctx context.Context, runID resource.ID, format PlanFormat) ([]byte, error) {
	subject, err := s.Authorize(ctx, authz.GetPlanFileAction, &authz.AccessRequest{ID: &runID})
	if err != nil {
		return nil, err
	}

	if plan, err := s.cache.Get(planFileCacheKey(format, runID)); err == nil {
		return plan, nil
	}
	// Cache is empty; retrieve from DB
	file, err := s.db.GetPlanFile(ctx, runID, format)
	if err != nil {
		s.Error(err, "retrieving plan file", "id", runID, "format", format, "subject", subject)
		return nil, err
	}
	// Cache plan before returning
	if err := s.cache.Set(planFileCacheKey(format, runID), file); err != nil {
		s.Error(err, "caching plan file")
	}
	return file, nil
}

// UploadPlanFile persists a run's plan file. The plan format should be either
// be binary or json.
func (s *Service) UploadPlanFile(ctx context.Context, runID resource.ID, plan []byte, format PlanFormat) error {
	subject, err := s.Authorize(ctx, authz.UploadPlanFileAction, &authz.AccessRequest{ID: &runID})
	if err != nil {
		return err
	}

	if err := s.db.SetPlanFile(ctx, runID, plan, format); err != nil {
		s.Error(err, "uploading plan file", "id", runID, "format", format, "subject", subject)
		return err
	}

	s.V(1).Info("uploaded plan file", "id", runID, "format", format, "subject", subject)

	if err := s.cache.Set(planFileCacheKey(format, runID), plan); err != nil {
		s.Error(err, "caching plan file")
	}

	return nil
}

// createReports creates reports of changes for the phase.
func (s *Service) createReports(ctx context.Context, runID resource.ID, phase internal.PhaseType) (resource Report, output Report, err error) {
	switch phase {
	case internal.PlanPhase:
		resource, output, err = s.createPlanReports(ctx, runID)
	case internal.ApplyPhase:
		resource, err = s.createApplyReport(ctx, runID)
	default:
		return Report{}, Report{}, fmt.Errorf("unknown supported phase for creating report: %s", phase)
	}
	return resource, output, err
}

func (s *Service) createPlanReports(ctx context.Context, runID resource.ID) (resources Report, outputs Report, err error) {
	plan, err := s.GetPlanFile(ctx, runID, PlanFormatJSON)
	if err != nil {
		return Report{}, Report{}, err
	}
	resourceReport, outputReport, err := CompilePlanReports(plan)
	if err != nil {
		return Report{}, Report{}, err
	}
	if err := s.db.CreatePlanReport(ctx, runID, resourceReport, outputReport); err != nil {
		return Report{}, Report{}, err
	}
	return resourceReport, outputReport, nil
}

func (s *Service) createApplyReport(ctx context.Context, runID resource.ID) (Report, error) {
	logs, err := s.logs.GetAllLogs(ctx, runID, internal.ApplyPhase)
	if err != nil {
		return Report{}, err
	}
	report, err := ParseApplyOutput(string(logs))
	if err != nil {
		return Report{}, err
	}
	if err := s.db.CreateApplyReport(ctx, runID, report); err != nil {
		return Report{}, err
	}
	return report, nil
}

func (s *Service) autoQueueRun(ctx context.Context, ws *workspace.Workspace) error {
	// Auto queue a run only if configured on the worspace and the workspace is
	// a connected to a VCS repo.
	if ws.QueueAllRuns && ws.Connection != nil {
		_, err := s.Create(ctx, ws.ID, CreateOptions{})
		if err != nil {
			return err
		}
	}
	return nil
}
