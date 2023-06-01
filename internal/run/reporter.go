package run

import (
	"context"
	"fmt"
	"net/url"

	"github.com/go-logr/logr"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/cloud"
	"github.com/leg100/otf/internal/configversion"
	"github.com/leg100/otf/internal/http/html/paths"
	"github.com/leg100/otf/internal/pubsub"
	"github.com/leg100/otf/internal/workspace"
)

// ReporterLockID is a unique ID guaranteeing only one reporter on a cluster is running at any time.
const ReporterLockID int64 = 179366396344335597

type (
	// Reporter reports back to VCS providers the current status of VCS-triggered
	// runs.
	Reporter struct {
		logr.Logger
		pubsub.Subscriber
		VCSProviderService
		ConfigurationVersionService
		WorkspaceService
		internal.HostnameService
	}

	ReporterOptions struct {
		ConfigurationVersionService configversion.Service
		WorkspaceService            workspace.Service
		VCSProviderService          VCSProviderService

		logr.Logger
		internal.DB
	}
)

// Start starts the reporter daemon. Should be invoked in a go routine.
func (r *Reporter) Start(ctx context.Context) error {
	// Unsubscribe whenever exiting this routine.
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
		if event.Type == pubsub.DeletedEvent {
			// Skip deleted run events
			continue
		}
		if err := r.handleRun(ctx, run); err != nil {
			return err
		}
	}
	return nil
}

func (r *Reporter) handleRun(ctx context.Context, run *Run) error {
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
	case internal.RunPending, internal.RunPlanQueued, internal.RunApplyQueued:
		status = cloud.VCSPendingStatus
	case internal.RunPlanning, internal.RunApplying, internal.RunPlanned, internal.RunConfirmed:
		status = cloud.VCSRunningStatus
	case internal.RunPlannedAndFinished:
		status = cloud.VCSSuccessStatus
		if run.Plan.ResourceReport != nil {
			description = fmt.Sprintf("planned: %s", run.Plan.ResourceReport)
		}
	case internal.RunApplied:
		status = cloud.VCSSuccessStatus
		if run.Apply.ResourceReport != nil {
			description = fmt.Sprintf("applied: %s", run.Apply.ResourceReport)
		}
	case internal.RunErrored, internal.RunCanceled, internal.RunForceCanceled, internal.RunDiscarded:
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
