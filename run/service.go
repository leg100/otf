package run

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/leg100/otf/configversion"
	"github.com/leg100/otf/logs"
	"github.com/leg100/otf/organization"
	"github.com/leg100/otf/pubsub"
	"github.com/leg100/otf/rbac"
	"github.com/leg100/otf/vcsprovider"
	"github.com/leg100/otf/workspace"
)

type (
	// Alias services so they don't conflict when nested together in struct
	RunService                  = Service
	ConfigurationVersionService configversion.Service
	WorkspaceService            workspace.Service
	VCSProviderService          vcsprovider.Service
	LogsService                 logs.Service

	Service interface {
		CreateRun(ctx context.Context, workspaceID string, opts RunCreateOptions) (*Run, error)
		GetRun(ctx context.Context, id string) (*Run, error)
		ListRuns(ctx context.Context, opts RunListOptions) (*RunList, error)

		EnqueuePlan(ctx context.Context, runID string) (*Run, error)

		// StartPhase starts a run phase.
		StartPhase(ctx context.Context, runID string, phase otf.PhaseType, _ PhaseStartOptions) (*Run, error)
		// FinishPhase finishes a phase. Creates a report of changes before updating the status of
		// the run.
		FinishPhase(ctx context.Context, runID string, phase otf.PhaseType, opts PhaseFinishOptions) (*Run, error)

		// GetPlanFile returns the plan file for the run.
		GetPlanFile(ctx context.Context, runID string, format PlanFormat) ([]byte, error)
		// UploadPlanFile persists a run's plan file. The plan format should be either
		// be binary or json.
		UploadPlanFile(ctx context.Context, runID string, plan []byte, format PlanFormat) error

		// Watch provides access to a stream of run events. The WatchOptions filters
		// events. Context must be cancelled to close stream.
		//
		// TODO(@leg100): it would be clearer to the caller if the stream is closed by
		// returning a stream object with a Close() method. The calling code would
		// call Watch(), and then defer a Close(), which is more readable IMO.
		Watch(ctx context.Context, opts WatchOptions) (<-chan otf.Event, error)

		get(ctx context.Context, runID string) (*Run, error)
		// apply enqueues an apply for the run.
		apply(ctx context.Context, runID string) error
		discard(ctx context.Context, runID string) error
		// cancel a run. If a run is in progress then a cancelation signal will be
		// sent out.
		cancel(ctx context.Context, runID string) error
		// forceCancel forcefully cancels a run.
		forceCancel(ctx context.Context, runID string) error

		// delete deletes a run.
		delete(ctx context.Context, runID string) error
		// createReport creates a report of changes for the phase.
		createReport(ctx context.Context, runID string, phase otf.PhaseType) (ResourceReport, error)
		createPlanReport(ctx context.Context, runID string) (ResourceReport, error)
		createApplyReport(ctx context.Context, runID string) (ResourceReport, error)

		lockFileService

		otf.Authorizer // run authorizer
		otf.Handlers   // http handlers
	}

	service struct {
		logr.Logger

		WorkspaceService
		otf.PubSubService

		site         otf.Authorizer
		organization otf.Authorizer
		workspace    otf.Authorizer
		*authorizer

		cache otf.Cache
		db    *pgdb
		*factory

		api *api
		web *webHandlers
	}

	Options struct {
		WorkspaceAuthorizer otf.Authorizer

		WorkspaceService
		ConfigurationVersionService

		logr.Logger
		otf.Cache
		otf.DB
		otf.Renderer
		*pubsub.Broker
		otf.Signer
	}
)

func NewService(opts Options) *service {
	db := newDB(opts.DB)
	svc := service{
		Logger:           opts.Logger,
		PubSubService:    opts.Broker,
		WorkspaceService: opts.WorkspaceService,
	}

	svc.site = &otf.SiteAuthorizer{opts.Logger}
	svc.organization = &organization.Authorizer{opts.Logger}
	svc.workspace = opts.WorkspaceAuthorizer
	svc.authorizer = &authorizer{db, opts.WorkspaceAuthorizer}

	svc.cache = opts.Cache
	svc.db = db
	svc.factory = &factory{
		opts.ConfigurationVersionService,
		opts.WorkspaceService,
	}

	svc.api = &api{
		Logger:           opts.Logger,
		jsonapiMarshaler: newJSONAPIMarshaler(opts.WorkspaceService, opts.Signer),
		svc:              &svc,
	}
	svc.web = &webHandlers{
		Logger:   opts.Logger,
		Renderer: opts.Renderer,
		svc:      &svc,
	}

	// Must register table name and service with pubsub broker so that it knows
	// how to lookup workspaces in the DB.
	opts.Register("runs", &svc)

	return &svc
}

func (s *service) AddHandlers(r *mux.Router) {
	s.api.addHandlers(r)
	s.web.addHandlers(r)
}

func (s *service) CreateRun(ctx context.Context, workspaceID string, opts RunCreateOptions) (*Run, error) {
	subject, err := s.workspace.CanAccess(ctx, rbac.CreateRunAction, workspaceID)
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

	s.Publish(otf.Event{Type: otf.EventRunCreated, Payload: run})

	return run, nil
}

func (s *service) GetRun(ctx context.Context, runID string) (*Run, error) {
	return s.get(ctx, runID)
}

// GetByID implements pubsub.Getter
func (s *service) GetByID(ctx context.Context, runID string) (any, error) {
	return s.db.GetRun(ctx, runID)
}

// ListRuns retrieves multiple runs. Use opts to filter and paginate the
// list.
func (s *service) ListRuns(ctx context.Context, opts RunListOptions) (*RunList, error) {
	var subject otf.Subject
	var authErr error
	if opts.Organization != nil && opts.WorkspaceName != nil {
		workspace, err := s.GetWorkspaceByName(ctx, *opts.Organization, *opts.WorkspaceName)
		if err != nil {
			return nil, err
		}
		// subject needs perms on workspace to list runs in workspace
		subject, authErr = s.workspace.CanAccess(ctx, rbac.GetWorkspaceAction, workspace.ID)
	} else if opts.WorkspaceID != nil {
		// subject needs perms on workspace to list runs in workspace
		subject, authErr = s.workspace.CanAccess(ctx, rbac.GetWorkspaceAction, *opts.WorkspaceID)
	} else if opts.Organization != nil {
		// subject needs perms on org to list runs in org
		subject, authErr = s.organization.CanAccess(ctx, rbac.ListRunsAction, *opts.Organization)
	} else {
		// subject needs to be site admin to list runs across site
		subject, authErr = s.site.CanAccess(ctx, rbac.ListRunsAction, "")
	}
	if authErr != nil {
		return nil, authErr
	}

	rl, err := s.db.ListRuns(ctx, opts)
	if err != nil {
		s.Error(err, "listing runs", "subject", subject)
		return nil, err
	}

	s.V(2).Info("listed runs", "count", len(rl.Items), "subject", subject)

	return rl, nil
}

// enqueuePlan enqueues a plan for the run.
//
// NOTE: this is an internal action, invoked by the scheduler only.
func (s *service) EnqueuePlan(ctx context.Context, runID string) (*Run, error) {
	subject, err := s.CanAccess(ctx, rbac.EnqueuePlanAction, runID)
	if err != nil {
		return nil, err
	}

	run, err := s.db.UpdateStatus(ctx, runID, func(run *Run) error {
		return run.EnqueuePlan()
	})
	if err != nil {
		s.Error(err, "enqueuing plan", "id", runID, "subject", subject)
		return nil, err
	}
	s.V(0).Info("enqueued plan", "id", runID, "subject", subject)

	s.Publish(otf.Event{Type: otf.EventRunStatusUpdate, Payload: run})

	return run, nil

}

func (s *service) Delete(ctx context.Context, runID string) error {
	return s.delete(ctx, runID)
}

// StartPhase starts a run phase.
func (s *service) StartPhase(ctx context.Context, runID string, phase otf.PhaseType, _ PhaseStartOptions) (*Run, error) {
	subject, err := s.CanAccess(ctx, rbac.StartPhaseAction, runID)
	if err != nil {
		return nil, err
	}

	run, err := s.db.UpdateStatus(ctx, runID, func(run *Run) error {
		return run.Start(phase)
	})
	if err != nil {
		s.Error(err, "starting "+string(phase), "id", runID, "subject", subject)
		return nil, err
	}
	s.V(0).Info("started "+string(phase), "id", runID, "subject", subject)
	s.Publish(otf.Event{Type: otf.EventRunStatusUpdate, Payload: run})
	return run, nil
}

// FinishPhase finishes a phase. Creates a report of changes before updating the status of
// the run.
func (s *service) FinishPhase(ctx context.Context, runID string, phase otf.PhaseType, opts PhaseFinishOptions) (*Run, error) {
	subject, err := s.CanAccess(ctx, rbac.FinishPhaseAction, runID)
	if err != nil {
		return nil, err
	}

	var report ResourceReport
	if !opts.Errored {
		var err error
		report, err = s.createReport(ctx, runID, phase)
		if err != nil {
			s.Error(err, "creating report", "id", runID, "phase", phase, "subject", subject)
			opts.Errored = true
		}
	}
	run, err := s.db.UpdateStatus(ctx, runID, func(run *Run) error {
		return run.Finish(phase, opts)
	})
	if err != nil {
		s.Error(err, "finishing "+string(phase), "id", runID, "subject", subject)
		return nil, err
	}
	s.V(0).Info("finished "+string(phase), "id", runID, "report", report, "subject", subject)
	s.Publish(otf.Event{Type: otf.EventRunStatusUpdate, Payload: run})
	return run, nil
}

// Watch provides authenticated access to a stream of run events.
func (s *service) Watch(ctx context.Context, opts WatchOptions) (<-chan otf.Event, error) {
	var err error
	if opts.WorkspaceID != nil {
		// caller must have workspace-level read permissions
		_, err = s.workspace.CanAccess(ctx, rbac.WatchAction, *opts.WorkspaceID)
	} else if opts.Organization != nil {
		// caller must have organization-level read permissions
		_, err = s.organization.CanAccess(ctx, rbac.WatchAction, *opts.Organization)
	} else {
		// caller must have site-level read permissions
		_, err = s.site.CanAccess(ctx, rbac.WatchAction, "")
	}
	if err != nil {
		return nil, err
	}

	if opts.Name == nil {
		opts.Name = otf.String("watch-" + otf.GenerateRandomString(6))
	}
	sub, err := s.Subscribe(ctx, *opts.Name)
	if err != nil {
		return nil, err
	}

	ch := make(chan otf.Event)
	go func() {
		for {
			select {
			case ev, ok := <-sub:
				if !ok {
					close(ch)
					return
				}

				run, ok := ev.Payload.(*Run)
				if !ok {
					continue // skip anything other than a run or a workspace
				}

				// apply workspace filter
				if opts.WorkspaceID != nil {
					if run.WorkspaceID != *opts.WorkspaceID {
						continue
					}
				}
				// apply organization filter
				if opts.Organization != nil {
					if run.Organization != *opts.Organization {
						continue
					}
				}

				ch <- ev
			case <-ctx.Done():
				close(ch)
				return
			}
		}
	}()
	return ch, nil
}

// GetRun retrieves a run from the db.
func (s *service) get(ctx context.Context, runID string) (*Run, error) {
	subject, err := s.CanAccess(ctx, rbac.GetRunAction, runID)
	if err != nil {
		return nil, err
	}

	run, err := s.db.GetRun(ctx, runID)
	if err != nil {
		s.Error(err, "retrieving run", "id", runID, "subject", subject)
		return nil, err
	}
	s.V(2).Info("retrieved run", "id", runID, "subject", subject)

	return run, nil
}

// apply enqueues an apply for the run.
func (s *service) apply(ctx context.Context, runID string) error {
	subject, err := s.CanAccess(ctx, rbac.ApplyRunAction, runID)
	if err != nil {
		return err
	}
	run, err := s.db.UpdateStatus(ctx, runID, func(run *Run) error {
		return run.EnqueueApply()
	})
	if err != nil {
		s.Error(err, "enqueuing apply", "id", runID, "subject", subject)
		return err
	}

	s.V(0).Info("enqueued apply", "id", runID, "subject", subject)

	s.Publish(otf.Event{Type: otf.EventRunStatusUpdate, Payload: run})

	return err
}

// discard discards the run.
func (s *service) discard(ctx context.Context, runID string) error {
	subject, err := s.CanAccess(ctx, rbac.DiscardRunAction, runID)
	if err != nil {
		return err
	}

	run, err := s.db.UpdateStatus(ctx, runID, func(run *Run) error {
		return run.Discard()
	})
	if err != nil {
		s.Error(err, "discarding run", "id", runID, "subject", subject)
		return err
	}

	s.V(0).Info("discarded run", "id", runID, "subject", subject)

	s.Publish(otf.Event{Type: otf.EventRunStatusUpdate, Payload: run})

	return err
}

// cancel a run. If a run is in progress then a cancelation signal will be
// sent out.
func (s *service) cancel(ctx context.Context, runID string) error {
	subject, err := s.CanAccess(ctx, rbac.CancelRunAction, runID)
	if err != nil {
		return err
	}

	var enqueue bool
	run, err := s.db.UpdateStatus(ctx, runID, func(run *Run) (err error) {
		enqueue, err = run.Cancel()
		return err
	})
	if err != nil {
		s.Error(err, "canceling run", "id", runID, "subject", subject)
		return err
	}
	s.V(0).Info("canceled run", "id", runID, "subject", subject)
	if enqueue {
		// notify agent which'll send a SIGINT to terraform
		s.Publish(otf.Event{Type: otf.EventRunCancel, Payload: run})
	}
	s.Publish(otf.Event{Type: otf.EventRunStatusUpdate, Payload: run})
	return nil
}

// ForceCancelRun forcefully cancels a run.
func (s *service) forceCancel(ctx context.Context, runID string) error {
	subject, err := s.CanAccess(ctx, rbac.CancelRunAction, runID)
	if err != nil {
		return err
	}
	run, err := s.db.UpdateStatus(ctx, runID, func(run *Run) error {
		return run.ForceCancel()
	})
	if err != nil {
		s.Error(err, "force canceling run", "id", runID, "subject", subject)
		return err
	}
	s.V(0).Info("force canceled run", "id", runID, "subject", subject)

	// notify agent which'll send a SIGKILL to terraform
	s.Publish(otf.Event{Type: otf.EventRunForceCancel, Payload: run})

	return err
}

func planFileCacheKey(f PlanFormat, id string) string {
	return fmt.Sprintf("%s.%s", id, f)
}

// GetPlanFile returns the plan file for the run.
func (s *service) GetPlanFile(ctx context.Context, runID string, format PlanFormat) ([]byte, error) {
	subject, err := s.CanAccess(ctx, rbac.GetPlanFileAction, runID)
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
		return nil, fmt.Errorf("caching plan: %w", err)
	}
	return file, nil
}

// UploadPlanFile persists a run's plan file. The plan format should be either
// be binary or json.
func (s *service) UploadPlanFile(ctx context.Context, runID string, plan []byte, format PlanFormat) error {
	subject, err := s.CanAccess(ctx, rbac.UploadPlanFileAction, runID)
	if err != nil {
		return err
	}

	if err := s.db.SetPlanFile(ctx, runID, plan, format); err != nil {
		s.Error(err, "uploading plan file", "id", runID, "format", format, "subject", subject)
		return err
	}

	s.V(1).Info("uploaded plan file", "id", runID, "format", format, "subject", subject)

	if err := s.cache.Set(planFileCacheKey(format, runID), plan); err != nil {
		return fmt.Errorf("caching plan: %w", err)
	}

	return nil
}

// delete a run.
func (s *service) delete(ctx context.Context, runID string) error {
	run, err := s.db.GetRun(ctx, runID)
	if err != nil {
		return err
	}

	subject, err := s.workspace.CanAccess(ctx, rbac.DeleteRunAction, run.WorkspaceID)
	if err != nil {
		return err
	}

	if err := s.db.DeleteRun(ctx, runID); err != nil {
		s.Error(err, "deleting run", "id", runID, "subject", subject)
		return err
	}
	s.V(0).Info("deleted run", "id", runID, "subject", subject)
	s.Publish(otf.Event{Type: otf.EventRunDeleted, Payload: run})
	return nil
}

// createReport creates a report of changes for the phase.
func (s *service) createReport(ctx context.Context, runID string, phase otf.PhaseType) (ResourceReport, error) {
	switch phase {
	case otf.PlanPhase:
		return s.createPlanReport(ctx, runID)
	case otf.ApplyPhase:
		return s.createApplyReport(ctx, runID)
	default:
		return ResourceReport{}, fmt.Errorf("unknown supported phase for creating report: %s", phase)
	}
}

func (s *service) createPlanReport(ctx context.Context, runID string) (ResourceReport, error) {
	plan, err := s.GetPlanFile(ctx, runID, PlanFormatJSON)
	if err != nil {
		return ResourceReport{}, err
	}
	report, err := CompilePlanReport(plan)
	if err != nil {
		return ResourceReport{}, err
	}
	if err := s.db.CreatePlanReport(ctx, runID, report); err != nil {
		return ResourceReport{}, err
	}
	return report, nil
}

func (s *service) createApplyReport(ctx context.Context, runID string) (ResourceReport, error) {
	logs, err := s.db.getLogs(ctx, runID, otf.ApplyPhase)
	if err != nil {
		return ResourceReport{}, err
	}
	report, err := ParseApplyOutput(string(logs))
	if err != nil {
		return ResourceReport{}, err
	}
	if err := s.db.CreateApplyReport(ctx, runID, report); err != nil {
		return ResourceReport{}, err
	}
	return report, nil
}
