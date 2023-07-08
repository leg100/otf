package run

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/configversion"
	"github.com/leg100/otf/internal/http/html"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/pubsub"
	"github.com/leg100/otf/internal/rbac"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/vcsprovider"
	"github.com/leg100/otf/internal/workspace"
)

type (
	// Alias services so they don't conflict when nested together in struct
	RunService                  = Service
	ConfigurationVersionService configversion.Service
	WorkspaceService            workspace.Service
	VCSProviderService          vcsprovider.Service

	Service interface {
		CreateRun(ctx context.Context, workspaceID string, opts RunCreateOptions) (*Run, error)
		GetRun(ctx context.Context, id string) (*Run, error)
		ListRuns(ctx context.Context, opts RunListOptions) (*resource.Page[*Run], error)
		EnqueuePlan(ctx context.Context, runID string) (*Run, error)
		// StartPhase starts a run phase.
		StartPhase(ctx context.Context, runID string, phase internal.PhaseType, _ PhaseStartOptions) (*Run, error)
		// FinishPhase finishes a phase. Creates a report of changes before updating the status of
		// the run.
		FinishPhase(ctx context.Context, runID string, phase internal.PhaseType, opts PhaseFinishOptions) (*Run, error)
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
		Watch(ctx context.Context, opts WatchOptions) (<-chan pubsub.Event, error)
		// Cancel a run. If a run is in progress then a cancelation signal will be
		// sent out.
		Cancel(ctx context.Context, runID string) (*Run, error)
		// Apply enqueues an Apply for the run.
		Apply(ctx context.Context, runID string) error
		// Delete a run.
		Delete(ctx context.Context, runID string) error

		// RetryRun retries a run, creating a new run with the same config
		// version.
		RetryRun(ctx context.Context, id string) (*Run, error)

		// DiscardRun discards a run. Run must be in the planned state.
		DiscardRun(ctx context.Context, runID string) error
		// ForceCancelRun forcefully cancels a run.
		ForceCancelRun(ctx context.Context, runID string) error

		lockFileService

		internal.Authorizer // run authorizer

		getLogs(ctx context.Context, runID string, phase internal.PhaseType) ([]byte, error)
	}

	service struct {
		logr.Logger

		WorkspaceService
		pubsub.PubSubService

		site         internal.Authorizer
		organization internal.Authorizer
		workspace    internal.Authorizer
		*authorizer

		cache internal.Cache
		db    *pgdb
		*factory

		web *webHandlers
	}

	Options struct {
		WorkspaceAuthorizer internal.Authorizer

		WorkspaceService
		ConfigurationVersionService
		VCSProviderService

		logr.Logger
		internal.Cache
		*sql.DB
		html.Renderer
		*pubsub.Broker
	}
)

func NewService(opts Options) *service {
	db := &pgdb{opts.DB}
	svc := service{
		Logger:           opts.Logger,
		PubSubService:    opts.Broker,
		WorkspaceService: opts.WorkspaceService,
	}

	svc.site = &internal.SiteAuthorizer{Logger: opts.Logger}
	svc.organization = &organization.Authorizer{Logger: opts.Logger}
	svc.workspace = opts.WorkspaceAuthorizer
	svc.authorizer = &authorizer{db, opts.WorkspaceAuthorizer}

	svc.cache = opts.Cache
	svc.db = db
	svc.factory = &factory{
		opts.ConfigurationVersionService,
		opts.WorkspaceService,
		opts.VCSProviderService,
	}

	svc.web = &webHandlers{
		Renderer:         opts.Renderer,
		WorkspaceService: opts.WorkspaceService,
		logger:           opts.Logger,
		svc:              &svc,
	}

	// Register with broker so that it can relay run events
	opts.Register("runs", &svc)

	return &svc
}

func (s *service) AddHandlers(r *mux.Router) {
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

	return run, nil
}

// GetRun retrieves a run from the db.
func (s *service) GetRun(ctx context.Context, runID string) (*Run, error) {
	subject, err := s.CanAccess(ctx, rbac.GetRunAction, runID)
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

// GetByID implements pubsub.Getter
func (s *service) GetByID(ctx context.Context, runID string, action pubsub.DBAction) (any, error) {
	if action == pubsub.DeleteDBAction {
		return &Run{ID: runID}, nil
	}
	return s.db.GetRun(ctx, runID)
}

// ListRuns retrieves multiple runs. Use opts to filter and paginate the
// list.
func (s *service) ListRuns(ctx context.Context, opts RunListOptions) (*resource.Page[*Run], error) {
	var (
		subject internal.Subject
		authErr error
	)
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

	page, err := s.db.ListRuns(ctx, opts)
	if err != nil {
		s.Error(err, "listing runs", "subject", subject)
		return nil, err
	}

	s.V(9).Info("listed runs", "count", len(page.Items), "subject", subject)

	return page, nil
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

	return run, nil
}

func (s *service) Delete(ctx context.Context, runID string) error {
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
	return nil
}

func (s *service) RetryRun(ctx context.Context, runID string) (*Run, error) {
	run, err := s.db.GetRun(ctx, runID)
	if err != nil {
		s.Error(err, "retrieving run", "id", runID)
		return nil, err
	}
	return s.CreateRun(ctx, run.WorkspaceID, RunCreateOptions{
		ConfigurationVersionID: &run.ConfigurationVersionID,
	})
}

// StartPhase starts a run phase.
func (s *service) StartPhase(ctx context.Context, runID string, phase internal.PhaseType, _ PhaseStartOptions) (*Run, error) {
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
	return run, nil
}

// FinishPhase finishes a phase. Creates a report of changes before updating the status of
// the run.
func (s *service) FinishPhase(ctx context.Context, runID string, phase internal.PhaseType, opts PhaseFinishOptions) (*Run, error) {
	subject, err := s.CanAccess(ctx, rbac.FinishPhaseAction, runID)
	if err != nil {
		return nil, err
	}

	var resourceReport, outputReport Report
	if !opts.Errored {
		var err error
		resourceReport, outputReport, err = s.createReports(ctx, runID, phase)
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
	s.V(0).Info("finished "+string(phase), "id", runID, "resource_changes", resourceReport, "output_changes", outputReport, "subject", subject)
	return run, nil
}

// Watch provides authenticated access to a stream of run events.
func (s *service) Watch(ctx context.Context, opts WatchOptions) (<-chan pubsub.Event, error) {
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

	sub, err := s.Subscribe(ctx, "run-watch-")
	if err != nil {
		return nil, err
	}

	// relay is returned to the caller to which filtered run events are sent
	relay := make(chan pubsub.Event)
	go func() {
		// relay events
		for ev := range sub {
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

			relay <- ev
		}
		close(relay)
	}()
	return relay, nil
}

// Apply enqueues an apply for the run.
func (s *service) Apply(ctx context.Context, runID string) error {
	subject, err := s.CanAccess(ctx, rbac.ApplyRunAction, runID)
	if err != nil {
		return err
	}
	_, err = s.db.UpdateStatus(ctx, runID, func(run *Run) error {
		return run.EnqueueApply()
	})
	if err != nil {
		s.Error(err, "enqueuing apply", "id", runID, "subject", subject)
		return err
	}

	s.V(0).Info("enqueued apply", "id", runID, "subject", subject)

	return err
}

// DiscardRun discards the run.
func (s *service) DiscardRun(ctx context.Context, runID string) error {
	subject, err := s.CanAccess(ctx, rbac.DiscardRunAction, runID)
	if err != nil {
		return err
	}

	_, err = s.db.UpdateStatus(ctx, runID, func(run *Run) error {
		return run.Discard()
	})
	if err != nil {
		s.Error(err, "discarding run", "id", runID, "subject", subject)
		return err
	}

	s.V(0).Info("discarded run", "id", runID, "subject", subject)

	return err
}

// Cancel a run. If a run is in progress then a cancelation signal will be
// sent out.
func (s *service) Cancel(ctx context.Context, runID string) (*Run, error) {
	subject, err := s.CanAccess(ctx, rbac.CancelRunAction, runID)
	if err != nil {
		return nil, err
	}

	run, err := s.db.UpdateStatus(ctx, runID, func(run *Run) (err error) {
		return run.Cancel()
	})
	if err != nil {
		s.Error(err, "canceling run", "id", runID, "subject", subject)
		return nil, err
	}
	s.V(0).Info("canceled run", "id", runID, "subject", subject)
	return run, nil
}

// ForceCancelRun forcefully cancels a run.
func (s *service) ForceCancelRun(ctx context.Context, runID string) error {
	subject, err := s.CanAccess(ctx, rbac.CancelRunAction, runID)
	if err != nil {
		return err
	}
	_, err = s.db.UpdateStatus(ctx, runID, func(run *Run) error {
		return run.ForceCancel()
	})
	if err != nil {
		s.Error(err, "force canceling run", "id", runID, "subject", subject)
		return err
	}
	s.V(0).Info("force canceled run", "id", runID, "subject", subject)

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
		s.Error(err, "caching plan file")
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
		s.Error(err, "caching plan file")
	}

	return nil
}

// createReports creates reports of changes for the phase.
func (s *service) createReports(ctx context.Context, runID string, phase internal.PhaseType) (resource Report, output Report, err error) {
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

func (s *service) createPlanReports(ctx context.Context, runID string) (resources Report, outputs Report, err error) {
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

func (s *service) createApplyReport(ctx context.Context, runID string) (Report, error) {
	logs, err := s.getLogs(ctx, runID, internal.ApplyPhase)
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

func (s *service) getLogs(ctx context.Context, runID string, phase internal.PhaseType) ([]byte, error) {
	data, err := s.db.Conn(ctx).FindLogs(ctx, sql.String(runID), sql.String(string(phase)))
	if err != nil {
		// Don't consider no rows an error because logs may not have been
		// uploaded yet.
		if sql.NoRowsInResultError(err) {
			return nil, nil
		}
		return nil, sql.Error(err)
	}
	return data, nil
}
