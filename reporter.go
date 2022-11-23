package otf

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/go-logr/logr"
	"gopkg.in/cenkalti/backoff.v1"
)

// ReporterLockID is shared by one or more reporters and is used to guarantee
// that only one reporter will run at any time.
const ReporterLockID int64 = 179366396344335597

// Reporter reports back to VCS providers the current status of VCS-triggered
// runs.
type Reporter struct {
	Application
	logr.Logger
	hostname string // otfd hostname
}

// NewReporter constructs and initialises the reporter.
func NewReporter(logger logr.Logger, app Application, hostname string) *Reporter {
	s := &Reporter{
		Application: app,
		Logger:      logger.WithValues("component", "reporter"),
		hostname:    hostname,
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

			cv, err := r.GetConfigurationVersion(ctx, run.ConfigurationVersionID())
			if err != nil {
				return err
			}

			// Skip runs that were not triggered via VCS
			if cv.IngressAttributes() == nil {
				continue
			}

			// need run's CV to determine if triggered via VCS
			// if not, skip
			// if so, then we need run's workspace's VCS provider to create a
			// VCS client
			// we then use VCS client to talk to relevant status API

			var status VCSStatus
			var description string
			switch run.Status() {
			case RunPending, RunPlanQueued, RunApplyQueued:
				status = VCSPendingStatus
			case RunPlanning, RunApplying, RunPlanned, RunConfirmed:
				status = VCSRunningStatus
			case RunPlannedAndFinished:
				status = VCSSuccessStatus
				if run.HasChanges() {
					description = fmt.Sprintf("planned: %s", run.Plan().ResourceReport)
				} else {
					description = "no changes"
				}
			case RunApplied:
				status = VCSSuccessStatus
				description = fmt.Sprintf("applied: %s", run.Apply().ResourceReport)
			case RunErrored, RunCanceled, RunForceCanceled, RunDiscarded:
				status = VCSErrorStatus
				description = run.Status().String()
			default:
				return fmt.Errorf("unknown run status: %s", run.Status())
			}

			ws, err := r.GetWorkspace(ctx, WorkspaceSpec{ID: String(run.WorkspaceID())})
			if err != nil {
				return err
			}

			if ws.Repo() == nil {
				return fmt.Errorf("workspace not connect to repo: %s", ws.ID())
			}

			err = r.SetStatus(ctx, ws.Repo().ProviderID, SetStatusOptions{
				Workspace:   ws.Name(),
				Ref:         cv.IngressAttributes().CommitSHA,
				Identifier:  ws.Repo().Identifier,
				Status:      status,
				Description: description,
				TargetURL: (&url.URL{
					Scheme: "https",
					Host:   r.hostname,
					Path:   RunGetPathUI(run),
				}).String(),
			})
			if err != nil {
				return err
			}
		}
	}
}

// ExclusiveReporter runs a reporter, ensuring it is the *only* reporter
// running.
func ExclusiveReporter(ctx context.Context, logger logr.Logger, hostname string, app LockableApplication) error {
	op := func() error {
		for {
			err := app.WithLock(ctx, ReporterLockID, func(app Application) error {
				return NewReporter(logger, app, hostname).Start(ctx)
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
