package run

import (
	"context"
	"fmt"
	"net/url"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
	"github.com/leg100/otf/cloud"
	"github.com/leg100/otf/configversion"
	"github.com/leg100/otf/http/html/paths"
	"github.com/leg100/otf/workspace"
	"gopkg.in/cenkalti/backoff.v1"
)

// reporterLockID is a unique ID guaranteeing only one reporter on a cluster is running at any time.
const reporterLockID int64 = 179366396344335597

type (
	// reporter reports back to VCS providers the current status of VCS-triggered
	// runs.
	reporter struct {
		logr.Logger
		otf.WatchService
		VCSProviderService
		ConfigurationVersionService
		WorkspaceService

		hostname string
	}

	ReporterOptions struct {
		ConfigurationVersionService configversion.Service
		WorkspaceService            workspace.Service
		VCSProviderService          VCSProviderService
		Hostname                    string

		logr.Logger
		otf.DB
		otf.WatchService
	}
)

// StartReporter starts a reporter.
func StartReporter(ctx context.Context, opts ReporterOptions) error {
	ctx = otf.AddSubjectToContext(ctx, &otf.Superuser{"reporter"})

	rptr := &reporter{
		Logger:                      opts.Logger.WithValues("component", "reporter"),
		WatchService:                opts.WatchService,
		hostname:                    opts.Hostname,
		ConfigurationVersionService: opts.ConfigurationVersionService,
		WorkspaceService:            opts.WorkspaceService,
	}

	op := func() error {
		// block on getting an exclusive lock
		lock, err := opts.WaitAndLock(ctx, reporterLockID)
		if err != nil {
			return err
		}
		defer lock.Release()

		rptr.V(2).Info("started")
		defer rptr.V(2).Info("stopped")

		err = rptr.start(ctx)
		select {
		case <-ctx.Done():
			return nil // exit
		default:
			return err // retry
		}
	}
	policy := backoff.WithContext(backoff.NewExponentialBackOff(), ctx)
	return backoff.RetryNotify(op, policy, nil)
}

// start starts the reporter daemon. Should be invoked in a go routine.
func (r *reporter) start(ctx context.Context) error {
	// Unsubscribe Watch() whenever exiting this routine.
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// subscribe to run events
	sub, err := r.Watch(ctx, otf.WatchOptions{Name: otf.String("reporter")})
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

func (r *reporter) handleRun(ctx context.Context, run *Run) error {
	cv, err := r.GetConfigurationVersion(ctx, run.ConfigurationVersionID)
	if err != nil {
		return err
	}

	// Skip runs that were not triggered via VCS
	if cv.IngressAttributes == nil {
		return nil
	}

	var status cloud.VCSStatus
	var description string
	switch run.Status {
	case otf.RunPending, otf.RunPlanQueued, otf.RunApplyQueued:
		status = cloud.VCSPendingStatus
	case otf.RunPlanning, otf.RunApplying, otf.RunPlanned, otf.RunConfirmed:
		status = cloud.VCSRunningStatus
	case otf.RunPlannedAndFinished:
		status = cloud.VCSSuccessStatus
		if run.Plan.ResourceReport != nil {
			description = fmt.Sprintf("planned: %s", run.Plan.ResourceReport)
		} else {
			description = "no changes"
		}
	case otf.RunApplied:
		status = cloud.VCSSuccessStatus
		if run.Plan.ResourceReport != nil {
			description = fmt.Sprintf("applied: %s", run.Plan.ResourceReport)
		} else {
			description = "no changes"
		}
	case otf.RunErrored, otf.RunCanceled, otf.RunForceCanceled, otf.RunDiscarded:
		status = cloud.VCSErrorStatus
		description = run.Status.String()
	default:
		return fmt.Errorf("unknown run status: %s", run.Status)
	}

	ws, err := r.GetWorkspace(ctx, run.WorkspaceID)
	if err != nil {
		return err
	}
	if ws.Connection == nil {
		return fmt.Errorf("workspace not connected to repo: %s", ws.ID)
	}

	client, err := r.GetVCSClient(ctx, ws.Connection.VCSProviderID)
	if err != nil {
		return err
	}

	return client.SetStatus(ctx, cloud.SetStatusOptions{
		Workspace:   ws.Name,
		Ref:         cv.IngressAttributes.CommitSHA,
		Repo:        cv.IngressAttributes.Repo,
		Status:      status,
		Description: description,
		TargetURL: (&url.URL{
			Scheme: "https",
			Host:   r.hostname,
			Path:   paths.Run(run.ID),
		}).String(),
	})
}
