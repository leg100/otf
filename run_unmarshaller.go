package otf

import (
	"encoding/json"
	"time"
)

type RunDBRow struct {
	RunID                 string                    `json:"run_id"`
	PlanID                string                    `json:"plan_id"`
	ApplyID               string                    `json:"apply_id"`
	CreatedAt             time.Time                 `json:"created_at"`
	UpdatedAt             time.Time                 `json:"updated_at"`
	IsDestroy             bool                      `json:"is_destroy"`
	PositionInQueue       int                       `json:"position_in_queue"`
	Refresh               bool                      `json:"refresh"`
	RefreshOnly           bool                      `json:"refresh_only"`
	Status                RunStatus                 `json:"status"`
	PlanStatus            PlanStatus                `json:"plan_status"`
	ApplyStatus           ApplyStatus               `json:"apply_status"`
	ReplaceAddrs          []string                  `json:"replace_addrs"`
	TargetAddrs           []string                  `json:"target_addrs"`
	PlanResourceReport    *ResourceReport           `json:"planned_changes"`
	ApplyResourceReport   *ResourceReport           `json:"applied_changes"`
	ConfigurationVersion  ConfigurationVersionDBRow `json:"configuration_version"`
	Workspace             WorkspaceDBRow            `json:"workspace"`
	RunStatusTimestamps   []RunStatusTimestamp      `json:"run_status_timestamps"`
	PlanStatusTimestamps  []PlanStatusTimestamp     `json:"plan_status_timestamps"`
	ApplyStatusTimestamps []ApplyStatusTimestamp    `json:"apply_status_timestamps"`
	FullCount             int                       `json:"full_count"`
}

func UnmarshalRunListFromDB(pgresult interface{}) (runs []*Run, count int, err error) {
	data, err := json.Marshal(pgresult)
	if err != nil {
		return nil, 0, err
	}
	var rows []RunDBRow
	if err := json.Unmarshal(data, &rows); err != nil {
		return nil, 0, err
	}

	for _, row := range rows {
		run, err := unmarshalRunDBRow(row)
		if err != nil {
			return nil, 0, err
		}
		runs = append(runs, run)
	}

	if len(rows) > 0 {
		count = rows[0].FullCount
	}

	return runs, count, nil
}

func UnmarshalRunFromDB(pgresult interface{}) (*Run, error) {
	data, err := json.Marshal(pgresult)
	if err != nil {
		return nil, err
	}
	var row RunDBRow
	if err := json.Unmarshal(data, &row); err != nil {
		return nil, err
	}

	return unmarshalRunDBRow(row)
}

func unmarshalRunDBRow(row RunDBRow) (*Run, error) {
	run := Run{
		ID: row.RunID,
		Timestamps: Timestamps{
			CreatedAt: row.CreatedAt,
			UpdatedAt: row.UpdatedAt,
		},
		IsDestroy:        row.IsDestroy,
		PositionInQueue:  row.PositionInQueue,
		Refresh:          row.Refresh,
		RefreshOnly:      row.RefreshOnly,
		Status:           row.Status,
		StatusTimestamps: row.RunStatusTimestamps,
		ReplaceAddrs:     row.ReplaceAddrs,
		TargetAddrs:      row.TargetAddrs,
		Plan: &Plan{
			ID:               row.PlanID,
			Status:           row.PlanStatus,
			ResourceReport:   row.PlanResourceReport,
			StatusTimestamps: row.PlanStatusTimestamps,
			RunID:            row.RunID,
		},
		Apply: &Apply{
			ID:               row.ApplyID,
			Status:           row.ApplyStatus,
			ResourceReport:   row.ApplyResourceReport,
			StatusTimestamps: row.ApplyStatusTimestamps,
			RunID:            row.RunID,
		},
	}

	var err error
	run.Workspace, err = UnmarshalWorkspaceFromDB(row.Workspace)
	if err != nil {
		return nil, err
	}

	run.ConfigurationVersion, err = UnmarshalConfigurationVersionFromDB(row.ConfigurationVersion)
	if err != nil {
		return nil, err
	}

	return &run, nil
}
