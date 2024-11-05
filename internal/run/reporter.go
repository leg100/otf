package run

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/configversion"
	"github.com/leg100/otf/internal/http/html/paths"
	"github.com/leg100/otf/internal/pubsub"
	"github.com/leg100/otf/internal/resource"
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
		*internal.HostnameService

		Configs    reporterConfigClient
		Workspaces reporterWorkspaceClient
		VCS        reporterVCSClient
		Runs       reporterRunClient

		// Cache most recently set status for each incomplete run to ensure the
		// same status is not set more than once on an upstream VCS provider.
		// This is important to avoid hitting rate limits on VCS providers, e.g.
		// GitHub has a limit of 1000 status updates on a commit:
		//
		// https://docs.github.com/en/rest/commits/statuses?apiVersion=2022-11-28#create-a-commit-status
		//
		// key is the run ID.
		Cache map[string]vcs.Status
	}

	reporterWorkspaceClient interface {
		Get(ctx context.Context, workspaceID resource.ID) (*workspace.Workspace, error)
	}

	reporterConfigClient interface {
		Get(ctx context.Context, id resource.ID) (*configversion.ConfigurationVersion, error)
	}

	reporterVCSClient interface {
		GetVCSClient(ctx context.Context, providerID resource.ID) (vcs.Client, error)
	}

	reporterRunClient interface {
		Watch(context.Context) (<-chan pubsub.Event[*Run], func())
	}
)

// Start starts the reporter daemon. Should be invoked in a go routine.
func (r *Reporter) Start(ctx context.Context) error {
	// subscribe to run events
	sub, unsub := r.Runs.Watch(ctx)
	defer unsub()

	for event := range sub {
		if event.Type == pubsub.DeletedEvent {
			// Skip deleted run events
			continue
		}
		if err := r.handleRun(ctx, event.Payload); err != nil {
			// any error is treated as non-fatal because reporting on runs is
			// considered "best-effort" rather than an integral operation
			r.Error(err, "reporting run vcs status", "run_id", event.Payload.ID)
			return nil
		}
	}
	return pubsub.ErrSubscriptionTerminated
}

func (r *Reporter) handleRun(ctx context.Context, run *Run) error {
	// Skip runs triggered via the UI or API
	if run.Source == SourceUI || run.Source == SourceAPI {
		return nil
	}

	cv, err := r.Configs.Get(ctx, run.ConfigurationVersionID)
	if err != nil {
		return err
	}

	// Skip runs with a configuration not sourced from a repo
	if cv.IngressAttributes == nil {
		return nil
	}

	ws, err := r.Workspaces.Get(ctx, run.WorkspaceID)
	if err != nil {
		return err
	}
	if ws.Connection == nil {
		return fmt.Errorf("workspace not connected to repo: %s", ws.ID)
	}

	client, err := r.VCS.GetVCSClient(ctx, ws.Connection.VCSProviderID)
	if err != nil {
		return err
	}

	// Report the status and description of the run state
	var (
		status      vcs.Status
		description string
	)
	switch run.Status {
	case RunPending, RunPlanQueued, RunApplyQueued, RunPlanning, RunApplying, RunPlanned, RunConfirmed:
		status = vcs.PendingStatus
	case RunPlannedAndFinished:
		status = vcs.SuccessStatus
		if run.Plan.ResourceReport != nil {
			description = fmt.Sprintf("planned: %s", run.Plan.ResourceReport)
		}
	case RunApplied:
		status = vcs.SuccessStatus
		if run.Apply.ResourceReport != nil {
			description = fmt.Sprintf("applied: %s", run.Apply.ResourceReport)
		}
	case RunErrored, RunCanceled, RunForceCanceled, RunDiscarded:
		status = vcs.ErrorStatus
		description = run.Status.String()
	default:
		return fmt.Errorf("unknown run status: %s", run.Status)
	}

	// Check status cache. If there is a hit for the same run and status then
	// skip setting the status again.
	if lastStatus, ok := r.Cache[run.ID]; ok && lastStatus == status {
		r.V(8).Info("skipped setting duplicate run status on vcs",
			"run_id", run.ID,
			"run_status", run.Status,
			"vcs_status", status,
		)
		return nil
	}

	err = client.SetStatus(ctx, vcs.SetStatusOptions{
		Workspace:   ws.Name,
		Ref:         cv.IngressAttributes.CommitSHA,
		Repo:        cv.IngressAttributes.Repo,
		Status:      status,
		Description: description,
		TargetURL:   r.URL(paths.Run(run.ID.String())),
	})
	if err != nil {
		return err
	}

	// Update status cache. If the run is complete then remove the run from the
	// cache because no further status updates are expected.
	if run.Done() {
		delete(r.Cache, run.ID)
	} else {
		r.Cache[run.ID] = status
	}

	return nil
}
