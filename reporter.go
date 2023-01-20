package otf

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/go-logr/logr"
	"github.com/leg100/otf/cloud"
	"github.com/leg100/otf/http/html/paths"
	"gopkg.in/cenkalti/backoff.v1"
)

// ReporterLockID is a unique ID guaranteeing only one reporter on a cluster is running at any time.
const ReporterLockID int64 = 179366396344335597

// Reporter reports back to VCS providers the current status of VCS-triggered
// runs.
type Reporter struct {
	Application
	logr.Logger
}

// NewReporter constructs and initialises the reporter.
func NewReporter(logger logr.Logger, app Application) *Reporter {
	s := &Reporter{
		Application: app,
		Logger:      logger.WithValues("component", "reporter"),
	}
	s.V(2).Info("started")

	return s
}

// Start starts the reporter daemon. Should be invoked in a go routine.
func (r *Reporter) Start(ctx context.Context) error {
	ctx = AddSubjectToContext(ctx, &Superuser{"reporter"})

	op := func() error {
		return r.reinitialize(ctx)
	}
	policy := backoff.WithContext(backoff.NewExponentialBackOff(), ctx)
	return backoff.RetryNotify(op, policy, func(err error, next time.Duration) {
		r.Error(err, "restarting reporter")
	})
}

// reinitialize retrieves workspaces and runs from the DB and listens to events,
// creating/deleting workspace queues accordingly and forwarding events to
// queues for scheduling.
func (r *Reporter) reinitialize(ctx context.Context) error {
	// Unsubscribe Watch() whenever exiting this routine.
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// subscribe to run events
	sub, err := r.Watch(ctx, WatchOptions{Name: String("reporter")})
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case event := <-sub:
			run, ok := event.Payload.(*Run)
			if !ok {
				// Skip non-run events
				continue
			}
			if err := r.handleRun(ctx, run); err != nil {
				return err
			}
		}
	}
}

// reinitialize retrieves workspaces and runs from the DB and listens to events,
// creating/deleting workspace queues accordingly and forwarding events to
// queues for scheduling.
func (r *Reporter) handleRun(ctx context.Context, run *Run) error {
	cv, err := r.GetConfigurationVersion(ctx, run.ConfigurationVersionID())
	if err != nil {
		return err
	}

	// Skip runs that were not triggered via VCS
	if cv.IngressAttributes() == nil {
		return nil
	}

	// need run's CV to determine if triggered via VCS
	// if not, skip
	// if so, then we need run's workspace's VCS provider to create a
	// VCS client
	// we then use VCS client to talk to relevant status API

	var status cloud.VCSStatus
	var description string
	switch run.Status() {
	case RunPending, RunPlanQueued, RunApplyQueued:
		status = cloud.VCSPendingStatus
	case RunPlanning, RunApplying, RunPlanned, RunConfirmed:
		status = cloud.VCSRunningStatus
	case RunPlannedAndFinished:
		status = cloud.VCSSuccessStatus
		if run.HasChanges() {
			description = fmt.Sprintf("planned: %s", run.Plan().ResourceReport)
		} else {
			description = "no changes"
		}
	case RunApplied:
		status = cloud.VCSSuccessStatus
		description = fmt.Sprintf("applied: %s", run.Apply().ResourceReport)
	case RunErrored, RunCanceled, RunForceCanceled, RunDiscarded:
		status = cloud.VCSErrorStatus
		description = run.Status().String()
	default:
		return fmt.Errorf("unknown run status: %s", run.Status())
	}

	ws, err := r.GetWorkspace(ctx, run.WorkspaceID())
	if err != nil {
		return err
	}
	if ws.Repo() == nil {
		return fmt.Errorf("workspace not connected to repo: %s", ws.ID())
	}

	client, err := r.GetVCSClient(ctx, ws.Repo().ProviderID)
	if err != nil {
		return err
	}

	return client.SetStatus(ctx, cloud.SetStatusOptions{
		Workspace:   ws.Name(),
		Ref:         cv.IngressAttributes().CommitSHA,
		Identifier:  ws.Repo().Identifier,
		Status:      status,
		Description: description,
		TargetURL: (&url.URL{
			Scheme: "https",
			Host:   r.Hostname(),
			Path:   paths.Run(run.ID()),
		}).String(),
	})
}

// ExclusiveReporter runs a reporter, ensuring it is the *only* reporter
// running.
func ExclusiveReporter(ctx context.Context, logger logr.Logger, hostname string, app LockableApplication) error {
	op := func() error {
		for {
			err := app.WithLock(ctx, ReporterLockID, func(app Application) error {
				return NewReporter(logger, app).Start(ctx)
			})
			select {
			case <-ctx.Done():
				return nil
			default:
				return err
			}
		}
	}
	return backoff.RetryNotify(op, backoff.NewExponentialBackOff(), nil)
}
