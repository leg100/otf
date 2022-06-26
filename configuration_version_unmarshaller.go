package otf

import (
	"github.com/jackc/pgtype"
	"github.com/leg100/otf/sql/pggen"
)

type ConfigurationVersionDBResult struct {
	ConfigurationVersionID               pgtype.Text                                  `json:"configuration_version_id"`
	CreatedAt                            pgtype.Timestamptz                           `json:"created_at"`
	AutoQueueRuns                        bool                                         `json:"auto_queue_runs"`
	Source                               pgtype.Text                                  `json:"source"`
	Speculative                          bool                                         `json:"speculative"`
	Status                               pgtype.Text                                  `json:"status"`
	WorkspaceID                          pgtype.Text                                  `json:"workspace_id"`
	ConfigurationVersionStatusTimestamps []pggen.ConfigurationVersionStatusTimestamps `json:"configuration_version_status_timestamps"`
}

func UnmarshalConfigurationVersionDBResult(result ConfigurationVersionDBResult) (*ConfigurationVersion, error) {
	cv := ConfigurationVersion{
		id:               result.ConfigurationVersionID.String,
		createdAt:        result.CreatedAt.Time,
		autoQueueRuns:    result.AutoQueueRuns,
		speculative:      result.Speculative,
		source:           ConfigurationSource(result.Source.String),
		status:           ConfigurationStatus(result.Status.String),
		statusTimestamps: unmarshalConfigurationVersionStatusTimestampDBTypes(result.ConfigurationVersionStatusTimestamps),
		workspaceID:      result.WorkspaceID.String,
	}
	return &cv, nil
}

func unmarshalConfigurationVersionStatusTimestampDBTypes(typs []pggen.ConfigurationVersionStatusTimestamps) (timestamps []ConfigurationVersionStatusTimestamp) {
	for _, ty := range typs {
		timestamps = append(timestamps, ConfigurationVersionStatusTimestamp{
			Status:    ConfigurationStatus(ty.Status.String),
			Timestamp: ty.Timestamp.Time,
		})
	}
	return timestamps
}
