package run

import (
	"context"
	"fmt"
	"net/url"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
	"github.com/leg100/otf/cloud"
	"github.com/leg100/otf/http/html/paths"
	"gopkg.in/cenkalti/backoff.v1"
)

// reporterLockID is a unique ID guaranteeing only one reporter on a cluster is running at any time.
const reporterLockID int64 = 179366396344335597

type (
	// reporter reports back to VCS providers the current status of VCS-triggered
	// runs.
	reporter struct {
		logr.Logger

		hostname string
		otf.ConfigurationVersionService
		otf.WatchService
		otf.WorkspaceService
		otf.VCSProviderService
	}

	ReporterOptions struct {
		logr.Logger
		otf.DB
		otf.HostnameService
		otf.ConfigurationVersionService
		otf.WatchService
		otf.WorkspaceService
		otf.VCSProviderService
	}
)

// StartReporter starts a reporter.
func StartReporter(ctx context.Context, opts ReporterOptions) error {
	ctx = otf.AddSubjectToContext(ctx, &otf.Superuser{"reporter"})

	op := func() error {
		// block on getting an exclusive lock
		lock, err := opts.WaitAndLock(ctx, reporterLockID)
		if err != nil {
			return err
		}
		defer lock.Release()

		s := &reporter{
			Logger:                      opts.Logger.WithValues("component", "reporter"),
			ConfigurationVersionService: opts.ConfigurationVersionService,
			WorkspaceService:            opts.WorkspaceService,
			WatchService:                opts.WatchService,
			hostname:                    opts.Hostname(),
		}
		s.V(2).Info("started")
		defer s.V(2).Info("stopped")

		err = s.start(ctx)
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
			run, ok := event.Payload.(*otf.Run)
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

func (r *reporter) handleRun(ctx context.Context, run *otf.Run) error {
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
	if ws.Repo == nil {
		return fmt.Errorf("workspace not connected to repo: %s", ws.ID)
	}

	client, err := r.GetVCSClient(ctx, ws.Repo.ProviderID)
	if err != nil {
		return err
	}

	return client.SetStatus(ctx, cloud.SetStatusOptions{
		Workspace:   ws.Name,
		Ref:         cv.IngressAttributes.CommitSHA,
		Identifier:  ws.Repo.Identifier,
		Status:      status,
		Description: description,
		TargetURL: (&url.URL{
			Scheme: "https",
			Host:   r.hostname,
			Path:   paths.Run(run.ID),
		}).String(),
	})
}
