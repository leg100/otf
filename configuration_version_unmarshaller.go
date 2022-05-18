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
		AutoQueueRuns:    result.AutoQueueRuns,
		Speculative:      result.Speculative,
		Source:           ConfigurationSource(result.Source),
		Status:           ConfigurationStatus(result.Status),
		StatusTimestamps: unmarshalConfigurationVersionStatusTimestampDBTypes(result.ConfigurationVersionStatusTimestamps),
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
		AutoQueueRuns: typ.AutoQueueRuns,
		Speculative:   typ.Speculative,
		Source:        ConfigurationSource(typ.Source),
		Status:        ConfigurationStatus(typ.Status),
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
