package otf

import (
	"time"

	"github.com/jackc/pgtype"
	"github.com/leg100/otf/sql/pggen"
)

// TODO: rename these objects from *DBResult to *DBRecord or just *Record

// RunDBResult is the database record for a run
type RunDBResult struct {
	RunID                  pgtype.Text                  `json:"run_id"`
	PlanID                 pgtype.Text                  `json:"plan_id"`
	ApplyID                pgtype.Text                  `json:"apply_id"`
	PlanJobID              pgtype.Text                  `json:"plan_job_id"`
	ApplyJobID             pgtype.Text                  `json:"apply_job_id"`
	CreatedAt              time.Time                    `json:"created_at"`
	IsDestroy              bool                         `json:"is_destroy"`
	PositionInQueue        int                          `json:"position_in_queue"`
	Refresh                bool                         `json:"refresh"`
	RefreshOnly            bool                         `json:"refresh_only"`
	Status                 pgtype.Text                  `json:"status"`
	PlanStatus             pgtype.Text                  `json:"plan_status"`
	ApplyStatus            pgtype.Text                  `json:"apply_status"`
	ReplaceAddrs           []string                     `json:"replace_addrs"`
	TargetAddrs            []string                     `json:"target_addrs"`
	PlannedAdditions       int                          `json:"planned_additions"`
	PlannedChanges         int                          `json:"planned_changes"`
	PlannedDestructions    int                          `json:"planned_destructions"`
	AppliedAdditions       int                          `json:"applied_additions"`
	AppliedChanges         int                          `json:"applied_changes"`
	AppliedDestructions    int                          `json:"applied_destructions"`
	ConfigurationVersionID pgtype.Text                  `json:"configuration_version_id"`
	WorkspaceID            pgtype.Text                  `json:"workspace_id"`
	Speculative            bool                         `json:"speculative"`
	AutoApply              bool                         `json:"auto_apply"`
	WorkspaceName          pgtype.Text                  `json:"workspace_name"`
	OrganizationName       pgtype.Text                  `json:"organization_name"`
	ConfigurationVersion   *pggen.ConfigurationVersions `json:"configuration_version"`
	Workspace              *pggen.Workspaces            `json:"workspace"`
	RunStatusTimestamps    []pggen.RunStatusTimestamps  `json:"run_status_timestamps"`
	PlanStatusTimestamps   []pggen.JobStatusTimestamps  `json:"plan_status_timestamps"`
	ApplyStatusTimestamps  []pggen.JobStatusTimestamps  `json:"apply_status_timestamps"`
}

func UnmarshalRunDBResult(result RunDBResult) (*Run, error) {
	run := Run{
		id:               result.RunID.String,
		createdAt:        result.CreatedAt,
		isDestroy:        result.IsDestroy,
		positionInQueue:  result.PositionInQueue,
		refresh:          result.Refresh,
		refreshOnly:      result.RefreshOnly,
		status:           RunStatus(result.Status.String),
		statusTimestamps: unmarshalRunStatusTimestampDBTypes(result.RunStatusTimestamps),
		replaceAddrs:     result.ReplaceAddrs,
		targetAddrs:      result.TargetAddrs,
		autoApply:        result.AutoApply,
		speculative:      result.Speculative,
		workspaceName:    result.WorkspaceName.String,
		organizationName: result.OrganizationName.String,
		Plan: &Plan{
			id: result.PlanID.String,
			job: &job{
				id:                     result.PlanJobID.String,
				runID:                  result.RunID.String,
				configurationVersionID: result.ConfigurationVersionID.String,
				workspaceID:            result.WorkspaceID.String,
				status:                 JobStatus(result.PlanStatus.String),
				statusTimestamps:       unmarshalJobStatusTimestampDBTypes(result.PlanStatusTimestamps),
				isDestroy:              result.IsDestroy,
			},
			ResourceReport: &ResourceReport{
				Additions:    result.PlannedAdditions,
				Changes:      result.PlannedChanges,
				Destructions: result.PlannedDestructions,
			},
		},
		Apply: &Apply{
			id: result.ApplyID.String,
			job: &job{
				id:                     result.ApplyJobID.String,
				runID:                  result.RunID.String,
				configurationVersionID: result.ConfigurationVersionID.String,
				workspaceID:            result.WorkspaceID.String,
				status:                 JobStatus(result.PlanStatus.String),
				statusTimestamps:       unmarshalJobStatusTimestampDBTypes(result.PlanStatusTimestamps),
				isDestroy:              result.IsDestroy,
			},
			ResourceReport: &ResourceReport{
				Additions:    result.AppliedAdditions,
				Changes:      result.AppliedChanges,
				Destructions: result.AppliedDestructions,
			},
		},
	}

	if result.Workspace != nil {
		workspace, err := UnmarshalWorkspaceDBType(*result.Workspace)
		if err != nil {
			return nil, err
		}
		run.Workspace = workspace
	} else {
		run.Workspace = &Workspace{id: result.WorkspaceID.String}
	}

	if result.ConfigurationVersion != nil {
		cv, err := unmarshalConfigurationVersionDBType(*result.ConfigurationVersion)
		if err != nil {
			return nil, err
		}
		run.ConfigurationVersion = cv
	} else {
		run.ConfigurationVersion = &ConfigurationVersion{id: result.ConfigurationVersionID.String}
	}

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
