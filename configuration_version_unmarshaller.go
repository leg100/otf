package otf

import (
	"time"

	"github.com/jackc/pgtype"
	"github.com/leg100/otf/sql/pggen"
)

type ConfigurationVersionDBResult struct {
	ConfigurationVersionID               pgtype.Text                                  `json:"configuration_version_id"`
	CreatedAt                            time.Time                                    `json:"created_at"`
	AutoQueueRuns                        bool                                         `json:"auto_queue_runs"`
	Source                               pgtype.Text                                  `json:"source"`
	Speculative                          bool                                         `json:"speculative"`
	Status                               pgtype.Text                                  `json:"status"`
	WorkspaceID                          pgtype.Text                                  `json:"workspace_id"`
	Workspace                            *pggen.Workspaces                            `json:"workspace"`
	ConfigurationVersionStatusTimestamps []pggen.ConfigurationVersionStatusTimestamps `json:"configuration_version_status_timestamps"`
}

func UnmarshalConfigurationVersionDBResult(result ConfigurationVersionDBResult) (*ConfigurationVersion, error) {
	cv := ConfigurationVersion{
		id:               result.ConfigurationVersionID.String,
		createdAt:        result.CreatedAt,
		autoQueueRuns:    result.AutoQueueRuns,
		speculative:      result.Speculative,
		source:           ConfigurationSource(result.Source.String),
		status:           ConfigurationStatus(result.Status.String),
		statusTimestamps: unmarshalConfigurationVersionStatusTimestampDBTypes(result.ConfigurationVersionStatusTimestamps),
	}

	if result.Workspace != nil {
		workspace, err := UnmarshalWorkspaceDBType(*result.Workspace)
		if err != nil {
			return nil, err
		}
		cv.Workspace = workspace
	} else {
		cv.Workspace = &Workspace{id: result.WorkspaceID.String}
	}

	return &cv, nil
}

func unmarshalConfigurationVersionDBType(typ pggen.ConfigurationVersions) (*ConfigurationVersion, error) {
	cv := ConfigurationVersion{
		id:            typ.ConfigurationVersionID.String,
		createdAt:     typ.CreatedAt.Local(),
		autoQueueRuns: typ.AutoQueueRuns,
		speculative:   typ.Speculative,
		source:        ConfigurationSource(typ.Source.String),
		status:        ConfigurationStatus(typ.Status.String),
		Workspace:     &Workspace{id: typ.WorkspaceID.String},
	}

	return &cv, nil
}

func unmarshalConfigurationVersionStatusTimestampDBTypes(typs []pggen.ConfigurationVersionStatusTimestamps) (timestamps []ConfigurationVersionStatusTimestamp) {
	for _, ty := range typs {
		timestamps = append(timestamps, ConfigurationVersionStatusTimestamp{
			Status:    ConfigurationStatus(ty.Status.String),
			Timestamp: ty.Timestamp.Local(),
		})
	}
	return timestamps
}
