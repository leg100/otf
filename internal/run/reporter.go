package run

import (
	"context"
	"fmt"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/configversion"
	"github.com/leg100/otf/internal/configversion/source"
	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/pubsub"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/runstatus"
	"github.com/leg100/otf/internal/ui/paths"
	"github.com/leg100/otf/internal/vcs"
	"github.com/leg100/otf/internal/workspace"
)

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
		// Key is the run ID.
		Cache map[resource.TfeID]vcs.Status
	}

	reporterWorkspaceClient interface {
		Get(ctx context.Context, workspaceID resource.TfeID) (*workspace.Workspace, error)
	}

	reporterConfigClient interface {
		Get(ctx context.Context, id resource.TfeID) (*configversion.ConfigurationVersion, error)
	}

	reporterVCSClient interface {
		Get(ctx context.Context, providerID resource.TfeID) (*vcs.Provider, error)
	}

	reporterRunClient interface {
		Watch(context.Context) (<-chan pubsub.Event[*Event], func())
		Get(context.Context, resource.TfeID) (*Run, error)
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

func (r *Reporter) handleRun(ctx context.Context, event *Event) error {
	// Skip runs triggered via the UI or API
	if event.Source == source.UI || event.Source == source.API {
		return nil
	}

	cv, err := r.Configs.Get(ctx, event.ConfigurationVersionID)
	if err != nil {
		return fmt.Errorf("retrieving config: %w", err)
	}

	// Skip runs with a configuration not sourced from a repo
	if cv.IngressAttributes == nil {
		return nil
	}

	ws, err := r.Workspaces.Get(ctx, event.WorkspaceID)
	if err != nil {
		return fmt.Errorf("retrieving workspace: %w", err)
	}
	if ws.Connection == nil {
		return fmt.Errorf("workspace not connected to repo: %s", ws.ID)
	}

	client, err := r.VCS.Get(ctx, ws.Connection.VCSProviderID)
	if err != nil {
		return fmt.Errorf("retrieving vcs client: %w", err)
	}

	run, err := r.Runs.Get(ctx, event.ID)
	if err != nil {
		return fmt.Errorf("retrieving run: %w", err)
	}

	// Report the status and description of the run state
	var (
		status      vcs.Status
		description string
	)
	switch event.Status {
	case runstatus.Pending, runstatus.PlanQueued, runstatus.ApplyQueued, runstatus.Planning, runstatus.Applying, runstatus.Planned, runstatus.Confirmed:
		status = vcs.PendingStatus
	case runstatus.PlannedAndFinished:
		status = vcs.SuccessStatus
		if run.Plan.ResourceReport != nil {
			description = fmt.Sprintf("planned: %s", run.Plan.ResourceReport)
		}
	case runstatus.Applied:
		status = vcs.SuccessStatus
		if run.Apply.ResourceReport != nil {
			description = fmt.Sprintf("applied: %s", run.Apply.ResourceReport)
		}
	case runstatus.Errored, runstatus.Canceled, runstatus.ForceCanceled, runstatus.Discarded:
		status = vcs.ErrorStatus
		description = event.Status.String()
	default:
		return fmt.Errorf("unknown run status: %s", event.Status)
	}

	// Check status cache. If there is a hit for the same run and status then
	// skip setting the status again.
	if lastStatus, ok := r.Cache[event.ID]; ok && lastStatus == status {
		r.V(8).Info("skipped setting duplicate run status on vcs",
			"run_id", event.ID,
			"run_status", event.Status,
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
		TargetURL:   r.URL(paths.Run(event.ID)),
	})
	if err != nil {
		return err
	}

	// Update status cache. If the run is complete then remove the run from the
	// cache because no further status updates are expected.
	if run.Done() {
		delete(r.Cache, event.ID)
	} else {
		r.Cache[event.ID] = status
	}

	return nil
}
