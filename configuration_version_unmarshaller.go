package otf

import (
	"time"

	"github.com/leg100/otf/sql/pggen"
)

type ConfigurationVersionDBResult struct {
	ConfigurationVersionID               string                                       `json:"configuration_version_id"`
	CreatedAt                            time.Time                                    `json:"created_at"`
	UpdatedAt                            time.Time                                    `json:"updated_at"`
	AutoQueueRuns                        bool                                         `json:"auto_queue_runs"`
	Source                               string                                       `json:"source"`
	Speculative                          bool                                         `json:"speculative"`
	Status                               string                                       `json:"status"`
	Workspace                            *pggen.Workspaces                            `json:"workspace"`
	ConfigurationVersionStatusTimestamps []pggen.ConfigurationVersionStatusTimestamps `json:"configuration_version_status_timestamps"`
}

func UnmarshalConfigurationVersionDBResult(result ConfigurationVersionDBResult) (*ConfigurationVersion, error) {
	cv := ConfigurationVersion{
		ID: result.ConfigurationVersionID,
		Timestamps: Timestamps{
			CreatedAt: result.CreatedAt,
			UpdatedAt: result.UpdatedAt,
		},
		autoQueueRuns:    result.AutoQueueRuns,
		speculative:      result.Speculative,
		source:           ConfigurationSource(result.Source),
		status:           ConfigurationStatus(result.Status),
		statusTimestamps: unmarshalConfigurationVersionStatusTimestampDBTypes(result.ConfigurationVersionStatusTimestamps),
	}

	workspace, err := unmarshalWorkspaceDBType(result.Workspace)
	if err != nil {
		return nil, err
	}
	cv.Workspace = workspace

	return &cv, nil
}

func unmarshalConfigurationVersionDBType(typ pggen.ConfigurationVersions) (*ConfigurationVersion, error) {
	cv := ConfigurationVersion{
		ID: typ.ConfigurationVersionID,
		Timestamps: Timestamps{
			CreatedAt: typ.CreatedAt.Local(),
			UpdatedAt: typ.UpdatedAt.Local(),
		},
		autoQueueRuns: typ.AutoQueueRuns,
		speculative:   typ.Speculative,
		source:        ConfigurationSource(typ.Source),
		status:        ConfigurationStatus(typ.Status),
		Workspace:     &Workspace{ID: typ.WorkspaceID},
	}

	return &cv, nil
}

func unmarshalConfigurationVersionStatusTimestampDBTypes(typs []pggen.ConfigurationVersionStatusTimestamps) (timestamps []ConfigurationVersionStatusTimestamp) {
	for _, ty := range typs {
		timestamps = append(timestamps, ConfigurationVersionStatusTimestamp{
			Status:    ConfigurationStatus(ty.Status),
			Timestamp: ty.Timestamp.Local(),
		})
	}
	return timestamps
}
