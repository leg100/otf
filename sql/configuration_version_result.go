package sql

import (
	"time"

	"github.com/leg100/otf"
)

type configurationVersionRow struct {
	ConfigurationVersionID               *string                                `json:"configuration_version_id"`
	CreatedAt                            time.Time                              `json:"created_at"`
	UpdatedAt                            time.Time                              `json:"updated_at"`
	AutoQueueRuns                        *bool                                  `json:"auto_queue_runs"`
	Source                               *string                                `json:"source"`
	Speculative                          *bool                                  `json:"speculative"`
	Status                               *string                                `json:"status"`
	Workspace                            Workspaces                             `json:"workspace"`
	ConfigurationVersionStatusTimestamps []ConfigurationVersionStatusTimestamps `json:"configuration_version_status_timestamps"`
}

func convertConfigurationVersion(result configurationVersionRow) *otf.ConfigurationVersion {
	cv := otf.ConfigurationVersion{}
	cv.ID = *result.ConfigurationVersionID
	cv.CreatedAt = result.CreatedAt
	cv.UpdatedAt = result.UpdatedAt
	cv.Status = otf.ConfigurationStatus(*result.Status)
	cv.Source = otf.ConfigurationSource(*result.Source)
	cv.AutoQueueRuns = *result.AutoQueueRuns
	cv.Speculative = *result.Speculative
	cv.Workspace = convertWorkspaceComposite(result.Workspace)

	for _, ts := range result.ConfigurationVersionStatusTimestamps {
		cv.StatusTimestamps = append(cv.StatusTimestamps, otf.ConfigurationVersionStatusTimestamp{
			Status:    otf.ConfigurationStatus(*ts.Status),
			Timestamp: ts.Timestamp,
		})
	}
	return &cv
}

func convertConfigurationVersionComposite(result ConfigurationVersions) *otf.ConfigurationVersion {
	cv := otf.ConfigurationVersion{}
	cv.ID = *result.ConfigurationVersionID
	cv.CreatedAt = result.CreatedAt
	cv.UpdatedAt = result.UpdatedAt
	cv.Status = otf.ConfigurationStatus(*result.Status)
	cv.Source = otf.ConfigurationSource(*result.Source)
	cv.AutoQueueRuns = *result.AutoQueueRuns
	cv.Speculative = *result.Speculative
	cv.Workspace = &otf.Workspace{
		ID: *result.WorkspaceID,
	}
	return &cv
}
