package otf

import (
	"time"

	"github.com/leg100/otf/sql/pggen"
)

// TODO: rename these objects from *DBResult to *DBRecord or just *Record

// RunDBResult is the database record for a run
type RunDBResult struct {
	RunID                  string                        `json:"run_id"`
	PlanID                 string                        `json:"plan_id"`
	ApplyID                string                        `json:"apply_id"`
	CreatedAt              time.Time                     `json:"created_at"`
	IsDestroy              bool                          `json:"is_destroy"`
	PositionInQueue        int                           `json:"position_in_queue"`
	Refresh                bool                          `json:"refresh"`
	RefreshOnly            bool                          `json:"refresh_only"`
	Status                 string                        `json:"status"`
	PlanStatus             string                        `json:"plan_status"`
	ApplyStatus            string                        `json:"apply_status"`
	ReplaceAddrs           []string                      `json:"replace_addrs"`
	TargetAddrs            []string                      `json:"target_addrs"`
	PlannedChanges         *pggen.ResourceReport         `json:"planned_changes"`
	AppliedChanges         *pggen.ResourceReport         `json:"applied_changes"`
	ConfigurationVersionID string                        `json:"configuration_version_id"`
	WorkspaceID            string                        `json:"workspace_id"`
	Speculative            bool                          `json:"speculative"`
	AutoApply              bool                          `json:"auto_apply"`
	ConfigurationVersion   *pggen.ConfigurationVersions  `json:"configuration_version"`
	Workspace              *pggen.Workspaces             `json:"workspace"`
	RunStatusTimestamps    []pggen.RunStatusTimestamps   `json:"run_status_timestamps"`
	PlanStatusTimestamps   []pggen.PlanStatusTimestamps  `json:"plan_status_timestamps"`
	ApplyStatusTimestamps  []pggen.ApplyStatusTimestamps `json:"apply_status_timestamps"`
}

func UnmarshalRunDBResult(result RunDBResult) (*Run, error) {
	run := Run{
		id:               result.RunID,
		createdAt:        result.CreatedAt,
		isDestroy:        result.IsDestroy,
		positionInQueue:  result.PositionInQueue,
		refresh:          result.Refresh,
		refreshOnly:      result.RefreshOnly,
		status:           RunStatus(result.Status),
		statusTimestamps: unmarshalRunStatusTimestampDBTypes(result.RunStatusTimestamps),
		replaceAddrs:     result.ReplaceAddrs,
		targetAddrs:      result.TargetAddrs,
		autoApply:        result.AutoApply,
		speculative:      result.Speculative,
		Plan: &Plan{
			id:               result.PlanID,
			status:           PlanStatus(result.PlanStatus),
			ResourceReport:   unmarshalResourceReportDBType(result.PlannedChanges),
			statusTimestamps: unmarshalPlanStatusTimestampDBTypes(result.PlanStatusTimestamps),
		},
		Apply: &Apply{
			id:               result.ApplyID,
			status:           ApplyStatus(result.ApplyStatus),
			ResourceReport:   unmarshalResourceReportDBType(result.AppliedChanges),
			statusTimestamps: unmarshalApplyStatusTimestampDBTypes(result.ApplyStatusTimestamps),
		},
	}
	run.Plan.run = &run
	run.Apply.run = &run
	run.setJob()

	if result.Workspace != nil {
		workspace, err := UnmarshalWorkspaceDBType(*result.Workspace)
		if err != nil {
			return nil, err
		}
		run.Workspace = workspace
	} else {
		run.Workspace = &Workspace{id: result.WorkspaceID}
	}

	if result.ConfigurationVersion != nil {
		cv, err := unmarshalConfigurationVersionDBType(*result.ConfigurationVersion)
		if err != nil {
			return nil, err
		}
		run.ConfigurationVersion = cv
	} else {
		run.ConfigurationVersion = &ConfigurationVersion{id: result.ConfigurationVersionID}
	}

	return &run, nil
}

func unmarshalResourceReportDBType(typ *pggen.ResourceReport) *ResourceReport {
	if typ == nil {
		return nil
	}

	return &ResourceReport{
		Additions:    typ.Additions,
		Changes:      typ.Changes,
		Destructions: typ.Destructions,
	}
}

func unmarshalRunStatusTimestampDBTypes(typs []pggen.RunStatusTimestamps) (timestamps []RunStatusTimestamp) {
	for _, ty := range typs {
		timestamps = append(timestamps, RunStatusTimestamp{
			Status:    RunStatus(ty.Status),
			Timestamp: ty.Timestamp.Local(),
		})
	}
	return timestamps
}

func unmarshalPlanStatusTimestampDBTypes(typs []pggen.PlanStatusTimestamps) (timestamps []PlanStatusTimestamp) {
	for _, ty := range typs {
		timestamps = append(timestamps, PlanStatusTimestamp{
			Status:    PlanStatus(ty.Status),
			Timestamp: ty.Timestamp.Local(),
		})
	}
	return timestamps
}

func unmarshalApplyStatusTimestampDBTypes(typs []pggen.ApplyStatusTimestamps) (timestamps []ApplyStatusTimestamp) {
	for _, ty := range typs {
		timestamps = append(timestamps, ApplyStatusTimestamp{
			Status:    ApplyStatus(ty.Status),
			Timestamp: ty.Timestamp.Local(),
		})
	}
	return timestamps
}
