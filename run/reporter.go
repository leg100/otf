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
		otf.Subscriber
		VCSProviderService
		ConfigurationVersionService
		WorkspaceService
		otf.HostnameService
	}

	ReporterOptions struct {
		ConfigurationVersionService configversion.Service
		WorkspaceService            workspace.Service
		VCSProviderService          VCSProviderService

		logr.Logger
		otf.DB
		otf.Subscriber
		otf.HostnameService
	}
)

// StartReporter starts a reporter.
func StartReporter(ctx context.Context, opts ReporterOptions) error {
	ctx = otf.AddSubjectToContext(ctx, &otf.Superuser{"reporter"})

	rptr := &reporter{
		Logger:                      opts.Logger.WithValues("component", "reporter"),
		Subscriber:                  opts.Subscriber,
		HostnameService:             opts.HostnameService,
		ConfigurationVersionService: opts.ConfigurationVersionService,
		WorkspaceService:            opts.WorkspaceService,
		VCSProviderService:          opts.VCSProviderService,
	}

	op := func() error {
		conn, err := opts.Acquire(ctx)
		if err != nil {
			return err
		}
		defer conn.Release()

		// block on getting an exclusive lock
		if _, err := conn.Exec(ctx, "SELECT pg_advisory_lock($1)", reporterLockID); err != nil {
			if ctx.Err() != nil {
				return nil // exit
			}
			return err // retry
		}

		return rptr.start(ctx)
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
	sub, err := r.Subscribe(ctx, "reporter-")
	if err != nil {
		return err
	}

	for event := range sub {
		run, ok := event.Payload.(*Run)
		if !ok {
			// Skip non-run events
			continue
		}
		if err := r.handleRun(ctx, run); err != nil {
			return err
		}
	}
	return nil
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

	var (
		status      cloud.VCSStatus
		description string
	)
	switch run.Status {
	case otf.RunPending, otf.RunPlanQueued, otf.RunApplyQueued:
		status = cloud.VCSPendingStatus
	case otf.RunPlanning, otf.RunApplying, otf.RunPlanned, otf.RunConfirmed:
		status = cloud.VCSRunningStatus
	case otf.RunPlannedAndFinished:
		status = cloud.VCSSuccessStatus
		if run.Plan.ResourceReport != nil {
			description = fmt.Sprintf("planned: %s", run.Plan.ResourceReport)
		}
	case otf.RunApplied:
		status = cloud.VCSSuccessStatus
		if run.Apply.ResourceReport != nil {
			description = fmt.Sprintf("applied: %s", run.Apply.ResourceReport)
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
			Host:   r.Hostname(),
			Path:   paths.Run(run.ID),
		}).String(),
	})
}
