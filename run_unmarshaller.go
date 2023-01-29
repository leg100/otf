package otf

import (
	"github.com/jackc/pgtype"
	"github.com/leg100/otf/http/dto"
	"github.com/leg100/otf/sql/pggen"
)

// RunResult represents the result of a database query for a run.
type RunResult struct {
	RunID                  pgtype.Text                   `json:"run_id"`
	CreatedAt              pgtype.Timestamptz            `json:"created_at"`
	ForceCancelAvailableAt pgtype.Timestamptz            `json:"force_cancel_available_at"`
	IsDestroy              bool                          `json:"is_destroy"`
	PositionInQueue        int                           `json:"position_in_queue"`
	Refresh                bool                          `json:"refresh"`
	RefreshOnly            bool                          `json:"refresh_only"`
	Status                 pgtype.Text                   `json:"status"`
	PlanStatus             pgtype.Text                   `json:"plan_status"`
	ApplyStatus            pgtype.Text                   `json:"apply_status"`
	ReplaceAddrs           []string                      `json:"replace_addrs"`
	TargetAddrs            []string                      `json:"target_addrs"`
	AutoApply              bool                          `json:"auto_apply"`
	PlannedChanges         *pggen.Report                 `json:"planned_changes"`
	AppliedChanges         *pggen.Report                 `json:"applied_changes"`
	ConfigurationVersionID pgtype.Text                   `json:"configuration_version_id"`
	WorkspaceID            pgtype.Text                   `json:"workspace_id"`
	Speculative            bool                          `json:"speculative"`
	ExecutionMode          pgtype.Text                   `json:"execution_mode"`
	Latest                 bool                          `json:"latest"`
	OrganizationName       pgtype.Text                   `json:"organization_name"`
	IngressAttributes      *pggen.IngressAttributes      `json:"ingress_attributes"`
	RunStatusTimestamps    []pggen.RunStatusTimestamps   `json:"run_status_timestamps"`
	PlanStatusTimestamps   []pggen.PhaseStatusTimestamps `json:"plan_status_timestamps"`
	ApplyStatusTimestamps  []pggen.PhaseStatusTimestamps `json:"apply_status_timestamps"`
}

func UnmarshalRunResult(result RunResult) (*Run, error) {
	run := Run{
		id:                     result.RunID.String,
		createdAt:              result.CreatedAt.Time.UTC(),
		isDestroy:              result.IsDestroy,
		positionInQueue:        result.PositionInQueue,
		refresh:                result.Refresh,
		refreshOnly:            result.RefreshOnly,
		status:                 RunStatus(result.Status.String),
		statusTimestamps:       unmarshalRunStatusTimestampRows(result.RunStatusTimestamps),
		replaceAddrs:           result.ReplaceAddrs,
		targetAddrs:            result.TargetAddrs,
		autoApply:              result.AutoApply,
		speculative:            result.Speculative,
		executionMode:          ExecutionMode(result.ExecutionMode.String),
		latest:                 result.Latest,
		organization:           result.OrganizationName.String,
		workspaceID:            result.WorkspaceID.String,
		configurationVersionID: result.ConfigurationVersionID.String,
		plan: &Plan{
			runID: result.RunID.String,
			phaseStatus: &phaseStatus{
				status:           PhaseStatus(result.PlanStatus.String),
				statusTimestamps: unmarshalPlanStatusTimestampRows(result.PlanStatusTimestamps),
			},
			ResourceReport: (*ResourceReport)(result.PlannedChanges),
		},
		apply: &Apply{
			runID: result.RunID.String,
			phaseStatus: &phaseStatus{
				status:           PhaseStatus(result.ApplyStatus.String),
				statusTimestamps: unmarshalApplyStatusTimestampRows(result.ApplyStatusTimestamps),
			},
			ResourceReport: (*ResourceReport)(result.AppliedChanges),
		},
	}
	if result.ForceCancelAvailableAt.Status == pgtype.Present {
		run.forceCancelAvailableAt = Time(result.ForceCancelAvailableAt.Time.UTC())
	}
	if result.IngressAttributes != nil {
		run.commit = &result.IngressAttributes.CommitSHA.String
	}
	return &run, nil
}

func UnmarshalRunJSONAPI(d *dto.Run) *Run {
	return &Run{
		id:                     d.ID,
		createdAt:              d.CreatedAt,
		forceCancelAvailableAt: d.ForceCancelAvailableAt,
		isDestroy:              d.IsDestroy,
		executionMode:          ExecutionMode(d.ExecutionMode),
		message:                d.Message,
		positionInQueue:        d.PositionInQueue,
		refresh:                d.Refresh,
		refreshOnly:            d.RefreshOnly,
		status:                 RunStatus(d.Status),
		// TODO: unmarshal timestamps
		replaceAddrs:           d.ReplaceAddrs,
		targetAddrs:            d.TargetAddrs,
		workspaceID:            d.Workspace.ID,
		configurationVersionID: d.ConfigurationVersion.ID,
		// TODO: unmarshal plan and apply relations
	}
}

// UnmarshalRunListJSONAPI converts a DTO into a run list
func UnmarshalRunListJSONAPI(json *dto.RunList) *RunList {
	wl := RunList{
		Pagination: UnmarshalPaginationJSONAPI(json.Pagination),
	}
	for _, i := range json.Items {
		wl.Items = append(wl.Items, UnmarshalRunJSONAPI(i))
	}

	return &wl
}

func unmarshalRunStatusTimestampRows(rows []pggen.RunStatusTimestamps) (timestamps []RunStatusTimestamp) {
	for _, ty := range rows {
		timestamps = append(timestamps, RunStatusTimestamp{
			Status:    RunStatus(ty.Status.String),
			Timestamp: ty.Timestamp.Time.UTC(),
		})
	}
	return timestamps
}

func unmarshalPlanStatusTimestampRows(rows []pggen.PhaseStatusTimestamps) (timestamps []PhaseStatusTimestamp) {
	for _, ty := range rows {
		timestamps = append(timestamps, PhaseStatusTimestamp{
			Status:    PhaseStatus(ty.Status.String),
			Timestamp: ty.Timestamp.Time.UTC(),
		})
	}
	return timestamps
}

func unmarshalApplyStatusTimestampRows(rows []pggen.PhaseStatusTimestamps) (timestamps []PhaseStatusTimestamp) {
	for _, ty := range rows {
		timestamps = append(timestamps, PhaseStatusTimestamp{
			Status:    PhaseStatus(ty.Status.String),
			Timestamp: ty.Timestamp.Time.UTC(),
		})
	}
	return timestamps
}
