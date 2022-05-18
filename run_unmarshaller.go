package otf

import (
	"time"

	"github.com/leg100/otf/sql/pggen"
)

type RunDBResult struct {
	RunID                 string                        `json:"run_id"`
	PlanID                string                        `json:"plan_id"`
	ApplyID               string                        `json:"apply_id"`
	CreatedAt             time.Time                     `json:"created_at"`
	UpdatedAt             time.Time                     `json:"updated_at"`
	IsDestroy             bool                          `json:"is_destroy"`
	PositionInQueue       int                           `json:"position_in_queue"`
	Refresh               bool                          `json:"refresh"`
	RefreshOnly           bool                          `json:"refresh_only"`
	Status                string                        `json:"status"`
	PlanStatus            string                        `json:"plan_status"`
	ApplyStatus           string                        `json:"apply_status"`
	ReplaceAddrs          []string                      `json:"replace_addrs"`
	TargetAddrs           []string                      `json:"target_addrs"`
	PlannedChanges        *pggen.ResourceReport         `json:"planned_changes"`
	AppliedChanges        *pggen.ResourceReport         `json:"applied_changes"`
	ConfigurationVersion  *pggen.ConfigurationVersions  `json:"configuration_version"`
	Workspace             *pggen.Workspaces             `json:"workspace"`
	RunStatusTimestamps   []pggen.RunStatusTimestamps   `json:"run_status_timestamps"`
	PlanStatusTimestamps  []pggen.PlanStatusTimestamps  `json:"plan_status_timestamps"`
	ApplyStatusTimestamps []pggen.ApplyStatusTimestamps `json:"apply_status_timestamps"`
}

func UnmarshalRunDBResult(result RunDBResult) (*Run, error) {
	run := Run{
		ID: result.RunID,
		Timestamps: Timestamps{
			CreatedAt: result.CreatedAt,
			UpdatedAt: result.UpdatedAt,
		},
		IsDestroy:        result.IsDestroy,
		PositionInQueue:  result.PositionInQueue,
		Refresh:          result.Refresh,
		RefreshOnly:      result.RefreshOnly,
		Status:           RunStatus(result.Status),
		StatusTimestamps: unmarshalRunStatusTimestampDBTypes(result.RunStatusTimestamps),
		ReplaceAddrs:     result.ReplaceAddrs,
		TargetAddrs:      result.TargetAddrs,
		Plan: &Plan{
			ID:               result.PlanID,
			Status:           PlanStatus(result.PlanStatus),
			ResourceReport:   unmarshalResourceReportDBType(result.PlannedChanges),
			StatusTimestamps: unmarshalPlanStatusTimestampDBTypes(result.PlanStatusTimestamps),
			RunID:            result.RunID,
		},
		Apply: &Apply{
			ID:               result.ApplyID,
			Status:           ApplyStatus(result.ApplyStatus),
			ResourceReport:   unmarshalResourceReportDBType(result.AppliedChanges),
			StatusTimestamps: unmarshalApplyStatusTimestampDBTypes(result.ApplyStatusTimestamps),
			RunID:            result.RunID,
		},
	}

	workspace, err := unmarshalWorkspaceDBType(result.Workspace)
	if err != nil {
		return nil, err
	}
	run.Workspace = workspace

	cv, err := unmarshalConfigurationVersionDBType(*result.ConfigurationVersion)
	if err != nil {
		return nil, err
	}
	run.ConfigurationVersion = cv

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
