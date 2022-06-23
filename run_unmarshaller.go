package otf

import (
	"time"

	"github.com/jackc/pgtype"
	"github.com/leg100/otf/sql/pggen"
)

// TODO: rename these objects from *DBResult to *DBRecord or just *Record

// RunDBResult is the database record for a run
type RunDBResult struct {
	RunID                  pgtype.Text                 `json:"run_id"`
	PlanID                 pgtype.Text                 `json:"plan_id"`
	ApplyID                pgtype.Text                 `json:"apply_id"`
	PlanJobID              pgtype.Text                 `json:"plan_job_id"`
	ApplyJobID             pgtype.Text                 `json:"apply_job_id"`
	CreatedAt              time.Time                   `json:"created_at"`
	IsDestroy              bool                        `json:"is_destroy"`
	PositionInQueue        int                         `json:"position_in_queue"`
	Refresh                bool                        `json:"refresh"`
	RefreshOnly            bool                        `json:"refresh_only"`
	Status                 pgtype.Text                 `json:"status"`
	PlanStatus             pgtype.Text                 `json:"plan_status"`
	ApplyStatus            pgtype.Text                 `json:"apply_status"`
	ReplaceAddrs           []string                    `json:"replace_addrs"`
	TargetAddrs            []string                    `json:"target_addrs"`
	PlannedChanges         *pggen.Report               `json:"planned_changes"`
	AppliedChanges         *pggen.Report               `json:"applied_changes"`
	ConfigurationVersionID pgtype.Text                 `json:"configuration_version_id"`
	WorkspaceID            pgtype.Text                 `json:"workspace_id"`
	Speculative            bool                        `json:"speculative"`
	AutoApply              bool                        `json:"auto_apply"`
	WorkspaceName          pgtype.Text                 `json:"workspace_name"`
	OrganizationName       pgtype.Text                 `json:"organization_name"`
	RunStatusTimestamps    []pggen.RunStatusTimestamps `json:"run_status_timestamps"`
	PlanStatusTimestamps   []pggen.JobStatusTimestamps `json:"plan_status_timestamps"`
	ApplyStatusTimestamps  []pggen.JobStatusTimestamps `json:"apply_status_timestamps"`
}

func UnmarshalRunDBResult(result RunDBResult, ws *Workspace) (*Run, error) {
	run := Run{
		id:                     result.RunID.String,
		createdAt:              result.CreatedAt,
		isDestroy:              result.IsDestroy,
		positionInQueue:        result.PositionInQueue,
		refresh:                result.Refresh,
		refreshOnly:            result.RefreshOnly,
		status:                 RunStatus(result.Status.String),
		statusTimestamps:       unmarshalRunStatusTimestampDBTypes(result.RunStatusTimestamps),
		replaceAddrs:           result.ReplaceAddrs,
		targetAddrs:            result.TargetAddrs,
		autoApply:              result.AutoApply,
		speculative:            result.Speculative,
		workspaceName:          result.WorkspaceName.String,
		organizationName:       result.OrganizationName.String,
		workspaceID:            result.WorkspaceID.String,
		workspace:              ws,
		configurationVersionID: result.ConfigurationVersionID.String,
		Plan: &Plan{
			id: result.PlanID.String,
			job: &job{
				id:               result.PlanJobID.String,
				status:           JobStatus(result.PlanStatus.String),
				statusTimestamps: unmarshalJobStatusTimestampDBTypes(result.PlanStatusTimestamps),
			},
			ResourceReport: (*ResourceReport)(result.PlannedChanges),
		},
		Apply: &Apply{
			id: result.ApplyID.String,
			job: &job{
				id:               result.ApplyJobID.String,
				status:           JobStatus(result.ApplyStatus.String),
				statusTimestamps: unmarshalJobStatusTimestampDBTypes(result.ApplyStatusTimestamps),
			},
			ResourceReport: (*ResourceReport)(result.AppliedChanges),
		},
	}
	run.Plan.run = &run
	run.Apply.run = &run
	run.setJob()
	return &run, nil
}

func unmarshalRunStatusTimestampDBTypes(typs []pggen.RunStatusTimestamps) (timestamps []RunStatusTimestamp) {
	for _, ty := range typs {
		timestamps = append(timestamps, RunStatusTimestamp{
			Status:    RunStatus(ty.Status.String),
			Timestamp: ty.Timestamp.Local(),
		})
	}
	return timestamps
}

func unmarshalJobStatusTimestampDBTypes(typs []pggen.JobStatusTimestamps) (timestamps []JobStatusTimestamp) {
	for _, ty := range typs {
		timestamps = append(timestamps, JobStatusTimestamp{
			Status:    JobStatus(ty.Status.String),
			Timestamp: ty.Timestamp.Local(),
		})
	}
	return timestamps
}
