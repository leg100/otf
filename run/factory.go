package run

import (
	"context"

	"github.com/leg100/otf"
	"github.com/leg100/otf/http/jsonapi"
)

// factory constructs runs
type factory struct {
	otf.ConfigurationVersionService
	otf.WorkspaceService
}

// NewRun constructs a new run at the beginning of its lifecycle using the
// provided options.
func (f *factory) NewRun(ctx context.Context, workspaceID string, opts RunCreateOptions) (*Run, error) {
	ws, err := f.GetWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	var cv otf.ConfigurationVersion
	if opts.ConfigurationVersionID != nil {
		cv, err = f.GetConfigurationVersion(ctx, *opts.ConfigurationVersionID)
	} else {
		cv, err = f.GetLatestConfigurationVersion(ctx, workspaceID)
	}
	if err != nil {
		return nil, err
	}

	return NewRun(cv, ws, opts), nil
}

// NewRun creates a new run with defaults.
func NewRun(cv otf.ConfigurationVersion, ws otf.Workspace, opts RunCreateOptions) *Run {
	run := Run{
		id:                     otf.NewID("run"),
		createdAt:              otf.CurrentTimestamp(),
		refresh:                DefaultRefresh,
		organization:           ws.Organization(),
		configurationVersionID: cv.ID(),
		workspaceID:            ws.ID(),
		speculative:            cv.Speculative(),
		replaceAddrs:           opts.ReplaceAddrs,
		targetAddrs:            opts.TargetAddrs,
		executionMode:          ws.ExecutionMode(),
		autoApply:              ws.AutoApply(),
	}
	run.plan = newPlan(&run)
	run.apply = newApply(&run)
	run.updateStatus(otf.RunPending)

	if opts.IsDestroy != nil {
		run.isDestroy = *opts.IsDestroy
	}
	if opts.Message != nil {
		run.message = *opts.Message
	}
	if opts.Refresh != nil {
		run.refresh = *opts.Refresh
	}
	if opts.AutoApply != nil {
		run.autoApply = *opts.AutoApply
	}
	if cv.IngressAttributes() != nil {
		run.commit = &cv.IngressAttributes().CommitSHA
	}
	return &run
}

// NewFromJSONAPI constructs a run from a json:api struct
func NewFromJSONAPI(d *jsonapi.Run) *Run {
	return &Run{
		id:                     d.ID,
		createdAt:              d.CreatedAt,
		forceCancelAvailableAt: d.ForceCancelAvailableAt,
		isDestroy:              d.IsDestroy,
		executionMode:          otf.ExecutionMode(d.ExecutionMode),
		message:                d.Message,
		positionInQueue:        d.PositionInQueue,
		refresh:                d.Refresh,
		refreshOnly:            d.RefreshOnly,
		status:                 otf.RunStatus(d.Status),
		// TODO: unmarshal timestamps
		replaceAddrs:           d.ReplaceAddrs,
		targetAddrs:            d.TargetAddrs,
		workspaceID:            d.Workspace.ID,
		configurationVersionID: d.ConfigurationVersion.ID,
		// TODO: unmarshal plan and apply relations
	}
}

// NewListFromJSONAPI constructs a run list from a json:api struct
func NewListFromJSONAPI(from *jsonapi.RunList) *otf.RunList {
	to := otf.RunList{
		Pagination: otf.NewPaginationFromJSONAPI(from.Pagination),
	}
	for _, i := range from.Items {
		to.Items = append(to.Items, NewFromJSONAPI(i))
	}
	return &to
}
