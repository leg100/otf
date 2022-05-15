package otf

import (
	"encoding/json"
	"time"
)

type ConfigurationVersionDBRow struct {
	ConfigurationVersionID               string                                `json:"configuration_version_id"`
	CreatedAt                            time.Time                             `json:"created_at"`
	UpdatedAt                            time.Time                             `json:"updated_at"`
	AutoQueueRuns                        bool                                  `json:"auto_queue_runs"`
	Source                               ConfigurationSource                   `json:"source"`
	Speculative                          bool                                  `json:"speculative"`
	Status                               ConfigurationStatus                   `json:"status"`
	Workspace                            *WorkspaceDBRow                       `json:"workspace"`
	WorkspaceID                          *string                               `json:"workspace_id"`
	ConfigurationVersionStatusTimestamps []ConfigurationVersionStatusTimestamp `json:"configuration_version_status_timestamps"`
	FullCount                            int                                   `json:"full_count"`
}

func UnmarshalConfigurationVersionListFromDB(pgresult interface{}) (cvs []*ConfigurationVersion, count int, err error) {
	data, err := json.Marshal(pgresult)
	if err != nil {
		return nil, 0, err
	}
	var rows []ConfigurationVersionDBRow
	if err := json.Unmarshal(data, &rows); err != nil {
		return nil, 0, err
	}

	for _, row := range rows {
		cv, err := unmarshalConfigurationVersionDBRow(row)
		if err != nil {
			return nil, 0, err
		}
		cvs = append(cvs, cv)
	}

	if len(rows) > 0 {
		count = rows[0].FullCount
	}

	return cvs, count, nil
}

func UnmarshalConfigurationVersionFromDB(pgresult interface{}) (*ConfigurationVersion, error) {
	data, err := json.Marshal(pgresult)
	if err != nil {
		return nil, err
	}
	var row ConfigurationVersionDBRow
	if err := json.Unmarshal(data, &row); err != nil {
		return nil, err
	}

	return unmarshalConfigurationVersionDBRow(row)
}

func unmarshalConfigurationVersionDBRow(row ConfigurationVersionDBRow) (*ConfigurationVersion, error) {
	cv := ConfigurationVersion{
		ID: row.ConfigurationVersionID,
		Timestamps: Timestamps{
			CreatedAt: row.CreatedAt,
			UpdatedAt: row.UpdatedAt,
		},
		AutoQueueRuns:    row.AutoQueueRuns,
		Speculative:      row.Speculative,
		Source:           row.Source,
		Status:           row.Status,
		StatusTimestamps: row.ConfigurationVersionStatusTimestamps,
	}

	if row.Workspace != nil {
		ws, err := UnmarshalWorkspaceFromDB(row.Workspace)
		if err != nil {
			return nil, err
		}
		cv.Workspace = ws
	}
	if row.WorkspaceID != nil {
		cv.Workspace = &Workspace{ID: *row.WorkspaceID}
	}

	return &cv, nil
}
