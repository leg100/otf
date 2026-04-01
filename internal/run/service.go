package run

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/a-h/templ"
	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/configversion"
	"github.com/leg100/otf/internal/configversion/source"
	"github.com/leg100/otf/internal/engine"
	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/pubsub"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/runstatus"
	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/user"
	"github.com/leg100/otf/internal/vcs"
	"github.com/leg100/otf/internal/workspace"
)

type (
	// Alias service to permit embedding it with other services in a struct
	// without a name clash.
	RunService = Service

	Service struct {
		logr.Logger
		authz.Interface
		*MetricsCollector

		client                 serviceClient
		db                     *pgdb
		afterCancelHooks       []func(context.Context, *Run) error
		afterForceCancelHooks  []func(context.Context, *Run) error
		afterEnqueuePlanHooks  []func(context.Context, *Run) error
		afterEnqueueApplyHooks []func(context.Context, *Run) error
		broker                 pubsub.SubscriptionService[*Event]
		tailer                 *tailer
		daemonCtx              context.Context

		*factory
	}

	Options struct {
		Authorizer         *authz.Authorizer
		VCSEventSubscriber vcs.Subscriber
		DaemonCtx          context.Context
		Client             serviceClient
		Logger             logr.Logger
		DB                 *sql.DB
		Listener           *sql.Listener
	}

	serviceClient interface {
		GetOrganization(ctx context.Context, name organization.Name) (*organization.Organization, error)
		GetWorkspace(context.Context, resource.ID) (*workspace.Workspace, error)
		GetWorkspaceByName(ctx context.Context, organization resource.ID, workspace string) (*workspace.Workspace, error)
		SetWorkspaceLatestRun(ctx context.Context, workspaceID, runID resource.ID) (*workspace.Workspace, error)
		Lock(ctx context.Context, workspaceID resource.ID, runID *resource.TfeID) (*workspace.Workspace, error)
		ListConnectedWorkspaces(ctx context.Context, vcsProviderID resource.ID, repoPath vcs.Repo) ([]*workspace.Workspace, error)
		AfterCreateWorkspace(hook func(context.Context, *workspace.Workspace) error)
		CreateConfigVersion(ctx context.Context, workspaceID resource.ID, opts configversion.CreateOptions) (*configversion.ConfigurationVersion, error)
		GetConfigVersion(ctx context.Context, id resource.ID) (*configversion.ConfigurationVersion, error)
		GetLatestConfigVersion(ctx context.Context, workspaceID resource.ID) (*configversion.ConfigurationVersion, error)
		UploadConfig(ctx context.Context, id resource.ID, config []byte) error
		GetSourceIcon(source source.Source) templ.Component
		GetVCSProvider(ctx context.Context, providerID resource.ID) (*vcs.Provider, error)
		GetLatest(ctx context.Context, engine *engine.Engine) (string, time.Time, error)
	}
)

func NewService(opts Options) *Service {
	db := &pgdb{opts.DB}
	svc := Service{
		Logger:    opts.Logger,
		client:    opts.Client,
		db:        db,
		Interface: opts.Authorizer,
		daemonCtx: opts.DaemonCtx,
	}
	svc.MetricsCollector = &MetricsCollector{
		service: &svc,
	}
	svc.factory = &factory{
		client: opts.Client,
	}
	svc.tailer = &tailer{
		broker: pubsub.NewBroker[Chunk](
			opts.Logger,
			opts.Listener,
			"chunks",
		),
		client: &svc,
	}
	spawner := &Spawner{
		Logger: opts.Logger.WithValues("component", "spawner"),
		client: struct {
			serviceClient
			*Service
		}{
			serviceClient: opts.Client,
			Service:       &svc,
		},
	}
	svc.broker = pubsub.NewBroker[*Event](
		opts.Logger,
		opts.Listener,
		"runs",
	)

	// Provide a means of looking up a run's parent workspace.
	opts.Authorizer.RegisterParentResolver(resource.RunKind,
		func(ctx context.Context, runID resource.ID) (resource.ID, error) {
			// NOTE: we look up directly in the database rather than via
			// service call to avoid a recursion loop.
			run, err := db.get(ctx, runID)
			if err != nil {
				return nil, err
			}
			return run.WorkspaceID, nil
		},
	)

	// Subscribe run spawner to incoming vcs events
	opts.VCSEventSubscriber.Subscribe(spawner.handle)

	// After a workspace is created, if auto-queue-runs is set, then create a
	// run as well.
	opts.Client.AfterCreateWorkspace(svc.autoQueueRun)

	return &svc
}

func (s *Service) CreateRun(ctx context.Context, workspaceID resource.ID, opts CreateOptions) (*Run, error) {
	subject, err := s.Authorize(ctx, authz.CreateRunAction, workspaceID)
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

// GetRun retrieves a run from the db.
func (s *Service) GetRun(ctx context.Context, runID resource.ID) (*Run, error) {
	subject, err := s.Authorize(ctx, authz.GetRunAction, runID)
	if err != nil {
		return nil, err
	}

	run, err := s.db.get(ctx, runID)
	if err != nil {
		s.Error(err, "retrieving run", "id", runID, "subject", subject)
		return nil, err
	}
	s.V(9).Info("retrieved run", "id", runID, "subject", subject)

	return run, nil
}

// ListRuns retrieves multiple runs. Use opts to filter and paginate the
// list.
func (s *Service) ListRuns(ctx context.Context, opts ListOptions) (*resource.Page[*Run], error) {
	var (
		subject authz.Subject
		authErr error
	)
	if opts.Organization != nil && opts.WorkspaceName != nil {
		workspace, err := s.client.GetWorkspaceByName(ctx, *opts.Organization, *opts.WorkspaceName)
		if err != nil {
			return nil, err
		}
		// subject needs perms on workspace to list runs in workspace
		subject, authErr = s.Authorize(ctx, authz.GetWorkspaceAction, workspace.ID)
	} else if opts.WorkspaceID != nil {
		// subject needs perms on workspace to list runs in workspace
		subject, authErr = s.Authorize(ctx, authz.GetWorkspaceAction, *opts.WorkspaceID)
	} else if opts.Organization != nil {
		// subject needs perms on org to list runs in org
		subject, authErr = s.Authorize(ctx, authz.ListRunsAction, *opts.Organization)
	} else {
		// subject needs to be site admin to list runs across site
		subject, authErr = s.Authorize(ctx, authz.ListRunsAction, resource.SiteID)
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

// ListRunsOlderThan lists runs created before t. Implements resource.deleterClient.
func (s *Service) ListRunsOlderThan(ctx context.Context, t time.Time) ([]*Run, error) {
	return resource.ListAll(func(opts resource.PageOptions) (*resource.Page[*Run], error) {
		return s.ListRuns(ctx, ListOptions{
			PageOptions:     opts,
			BeforeCreatedAt: &t,
		})
	})
}

func (s *Service) listStatuses(ctx context.Context) ([]status, error) {
	return s.db.listStatuses(ctx)
}

// EnqueuePlan enqueues a plan for the run.
func (s *Service) EnqueuePlan(ctx context.Context, runID resource.ID) (run *Run, err error) {
	err = s.db.Tx(ctx, func(ctx context.Context) error {
		run, err = s.db.UpdateStatus(ctx, runID, func(ctx context.Context, run *Run) error {
			return run.EnqueuePlan()
		})
		if err != nil {
			return err
		}
		if !run.PlanOnly {
			_, err := s.client.Lock(ctx, run.WorkspaceID, &run.ID)
			if err != nil {
				return err
			}
			_, err = s.client.SetWorkspaceLatestRun(ctx, run.WorkspaceID, run.ID)
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

func (s *Service) DeleteRun(ctx context.Context, runID resource.ID) error {
	subject, err := s.Authorize(ctx, authz.DeleteRunAction, runID)
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
func (s *Service) StartPhase(ctx context.Context, runID resource.ID, phase PhaseType, _ PhaseStartOptions) (*Run, error) {
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
func (s *Service) FinishPhase(ctx context.Context, runID resource.ID, phase PhaseType, opts PhaseFinishOptions) (*Run, error) {
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
	err := s.db.Tx(ctx, func(ctx context.Context) (err error) {
		var autoapply bool
		run, err = s.db.UpdateStatus(ctx, runID, func(ctx context.Context, run *Run) (err error) {
			autoapply, err = run.Finish(phase, opts)
			return err
		})
		if err != nil {
			return err
		}
		if autoapply {
			return s.ApplyRun(ctx, runID)
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

func (s *Service) WatchRuns(ctx context.Context) (<-chan pubsub.Event[*Event], func()) {
	return s.broker.Subscribe(ctx)
}

// ApplyRun enqueues an apply for the run.
func (s *Service) ApplyRun(ctx context.Context, runID resource.ID) error {
	subject, err := s.Authorize(ctx, authz.ApplyRunAction, runID)
	if err != nil {
		return err
	}
	return s.db.Tx(ctx, func(ctx context.Context) error {
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

// DiscardRun discards the run.
func (s *Service) DiscardRun(ctx context.Context, runID resource.ID) error {
	subject, err := s.Authorize(ctx, authz.DiscardRunAction, runID)
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

func (s *Service) CancelRun(ctx context.Context, runID resource.ID) error {
	subject, err := s.Authorize(ctx, authz.CancelRunAction, runID)
	if err != nil {
		return err
	}
	var run *Run
	err = s.db.Tx(ctx, func(ctx context.Context) error {
		_, isUser := subject.(*user.User)

		var err error
		run, err = s.db.UpdateStatus(ctx, runID, func(ctx context.Context, run *Run) (err error) {
			return run.Cancel(isUser, false)
		})
		if err != nil {
			return err
		}
		// invoke AfterCancel hooks
		for _, hook := range s.afterCancelHooks {
			if err := hook(ctx, run); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		s.Error(err, "canceling run", "id", runID, "subject", subject)
	}
	if run.Status != runstatus.Canceled && run.CancelSignaledAt != nil {
		s.V(0).Info("signaled cancelation", "id", runID, "subject", subject)

		// After the cool off period, send an event, which'll refresh the UI to
		// inform the user the run can be force canceled.
		go func() {
			select {
			case <-s.daemonCtx.Done():
				return
			case <-time.After(forceCancelCoolOff):
			}
			if err := s.db.triggerEvent(s.daemonCtx, run.ID); err != nil {
				s.Error(err, "updating run after for cancel cool off period", "run", run)
			}
		}()
	} else {
		s.V(0).Info("canceled run", "id", runID, "subject", subject)
	}
	return nil
}

func (s *Service) AfterCancelRun(hook func(context.Context, *Run) error) {
	// add hook to list of hooks to be triggered after run is canceled
	s.afterCancelHooks = append(s.afterCancelHooks, hook)
}

// ForceCancelRun forcefully cancels a run.
func (s *Service) ForceCancelRun(ctx context.Context, runID resource.ID) error {
	subject, err := s.Authorize(ctx, authz.ForceCancelRunAction, runID)
	if err != nil {
		return err
	}
	return s.db.Tx(ctx, func(ctx context.Context) error {
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

// GetRunPlanFile returns the plan file for the run.
func (s *Service) GetRunPlanFile(ctx context.Context, runID resource.ID, format PlanFormat) ([]byte, error) {
	subject, err := s.Authorize(ctx, authz.GetPlanFileAction, runID)
	if err != nil {
		return nil, err
	}

	file, err := s.db.GetPlanFile(ctx, runID, format)
	if err != nil {
		s.Error(err, "retrieving plan file", "id", runID, "format", format, "subject", subject)
		return nil, err
	}
	return file, nil
}

// UploadRunPlanFile persists a run's plan file. The plan format should be either
// be binary or json.
func (s *Service) UploadRunPlanFile(ctx context.Context, runID resource.ID, plan []byte, format PlanFormat) error {
	subject, err := s.Authorize(ctx, authz.UploadPlanFileAction, runID)
	if err != nil {
		return err
	}

	if err := s.db.SetPlanFile(ctx, runID, plan, format); err != nil {
		s.Error(err, "uploading plan file", "id", runID, "format", format, "subject", subject)
		return err
	}

	s.V(1).Info("uploaded plan file", "id", runID, "format", format, "subject", subject)

	return nil
}

// createReports creates reports of changes for the phase.
func (s *Service) createReports(ctx context.Context, runID resource.ID, phase PhaseType) (resource Report, output Report, err error) {
	switch phase {
	case PlanPhase:
		resource, output, err = s.createPlanReports(ctx, runID)
	case ApplyPhase:
		resource, err = s.createApplyReport(ctx, runID)
	default:
		return Report{}, Report{}, fmt.Errorf("unknown supported phase for creating report: %s", phase)
	}
	return resource, output, err
}

func (s *Service) createPlanReports(ctx context.Context, runID resource.ID) (resources Report, outputs Report, err error) {
	plan, err := s.GetRunPlanFile(ctx, runID, PlanFormatJSON)
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
	logs, err := s.GetChunk(ctx, GetChunkOptions{
		RunID: runID.(resource.TfeID),
		Phase: ApplyPhase,
	})
	if err != nil {
		return Report{}, err
	}
	report, err := ParseApplyOutput(string(logs.Data))
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
		_, err := s.CreateRun(ctx, ws.ID, CreateOptions{})
		if err != nil {
			return err
		}
	}
	return nil
}

//
// Run log stuff
//

// GetChunk retrieves a chunk of logs for a run phase.
func (s *Service) GetChunk(ctx context.Context, opts GetChunkOptions) (Chunk, error) {
	chunk, err := s.db.getChunk(ctx, opts)
	if err != nil {
		s.Error(err, "retrieving log chunk", "run_id", opts.RunID, "phase", opts.Phase, "offset", opts.Offset, "limit", opts.Limit)
		return Chunk{}, err
	}
	s.V(9).Info("retrieved log chunk", "chunk", chunk)
	return chunk, nil
}

// PutChunk writes a chunk of logs for a run phase
func (s *Service) PutChunk(ctx context.Context, opts PutChunkOptions) error {
	_, err := s.Authorize(ctx, authz.PutChunkAction, opts.RunID)
	if err != nil {
		return err
	}

	chunk, err := newChunk(opts)
	if err != nil {
		s.Error(err, "creating log chunk", "run_id", opts, "phase", opts.Phase, "offset", opts.Offset)
		return err
	}
	if err := s.db.putChunk(ctx, chunk); err != nil {
		s.Error(err, "writing log chunk", "chunk", chunk)
		return err
	}
	s.V(3).Info("written log chunk", "chunk", chunk)

	return nil
}

// TailRun tails logs for a phase. Offset specifies the number of bytes into the logs
// from which to start tailing.
func (s *Service) TailRun(ctx context.Context, opts TailOptions) (<-chan Chunk, error) {
	subject, err := s.Authorize(ctx, authz.TailLogsAction, opts.RunID)
	if err != nil {
		return nil, err
	}
	tail, err := s.tailer.Tail(ctx, opts)
	if err != nil {
		s.Error(err, "tailing logs", "id", opts.RunID, "offset", opts.Offset, "subject", subject)
		return nil, err
	}
	s.V(9).Info("tailing logs", "id", opts.RunID, "phase", opts.Phase, "subject", subject)
	return tail, nil
}
