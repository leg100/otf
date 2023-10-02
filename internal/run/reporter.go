package run

import (
	"context"
	"fmt"
	"net/url"

	"github.com/go-logr/logr"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/configversion"
	"github.com/leg100/otf/internal/http/html/paths"
	"github.com/leg100/otf/internal/pubsub"
	"github.com/leg100/otf/internal/vcs"
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
	return pubsub.ErrSubscriptionTerminated
}

func (r *Reporter) handleRun(ctx context.Context, run *Run) error {
	// Skip runs triggered via the UI
	if run.Source == SourceUI {
		return nil
	}

	cv, err := r.GetConfigurationVersion(ctx, run.ConfigurationVersionID)
	if err != nil {
		return err
	}

	// Skip runs with a configuration not sourced from a repo
	if cv.IngressAttributes == nil {
		return nil
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

	// Report the status and description of the run state
	var (
		status      vcs.Status
		description string
	)
	switch run.Status {
	case internal.RunPending, internal.RunPlanQueued, internal.RunApplyQueued:
		status = vcs.PendingStatus
	case internal.RunPlanning, internal.RunApplying, internal.RunPlanned, internal.RunConfirmed:
		status = vcs.RunningStatus
	case internal.RunPlannedAndFinished:
		status = vcs.SuccessStatus
		if run.Plan.ResourceReport != nil {
			description = fmt.Sprintf("planned: %s", run.Plan.ResourceReport)
		}
	case internal.RunApplied:
		status = vcs.SuccessStatus
		if run.Apply.ResourceReport != nil {
			description = fmt.Sprintf("applied: %s", run.Apply.ResourceReport)
		}
	case internal.RunErrored, internal.RunCanceled, internal.RunForceCanceled, internal.RunDiscarded:
		status = vcs.ErrorStatus
		description = run.Status.String()
	default:
		return fmt.Errorf("unknown run status: %s", run.Status)
	}
	return client.SetStatus(ctx, vcs.SetStatusOptions{
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
