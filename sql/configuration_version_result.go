package sql

import (
	"time"

	"github.com/leg100/otf"
)

type configurationVersionResultWithoutRelations interface {
	GetConfigurationVersionID() *string
	GetAutoQueueRuns() *bool
	GetSource() *string
	GetSpeculative() *bool
	GetStatus() *string

	Timestamps
}

type configurationVersionResult interface {
	configurationVersionResultWithoutRelations

	GetConfigurationVersionStatusTimestamps() []ConfigurationVersionStatusTimestamps
	GetWorkspace() Workspaces
}

type configurationVersionListResult interface {
	configurationVersionResult

	GetFullCount() *int
}

type configurationVersionStatusTimestamp interface {
	GetConfigurationVersionID() *string
	GetStatus() *string
	GetTimestamp() time.Time
}

func addResultToConfigurationVersion(cv *otf.ConfigurationVersion, result configurationVersionResultWithoutRelations) {
	cv.ID = *result.GetConfigurationVersionID()
	cv.Timestamps = convertTimestamps(result)
	cv.Status = otf.ConfigurationStatus(*result.GetStatus())
	cv.Source = otf.ConfigurationSource(*result.GetSource())
	cv.AutoQueueRuns = *result.GetAutoQueueRuns()
	cv.Speculative = *result.GetSpeculative()
}

func convertConfigurationVersionResult(row configurationVersionResultWithoutRelations) *otf.ConfigurationVersion {
	cv := otf.ConfigurationVersion{}
	addResultToConfigurationVersion(&cv, row)
	return &cv
}

func convertConfigurationVersion(result configurationVersionResult) *otf.ConfigurationVersion {
	cv := convertConfigurationVersionResult(result)
	cv.Workspace = convertWorkspaceComposite(result.GetWorkspace())

	for _, ts := range result.GetConfigurationVersionStatusTimestamps() {
		cv.StatusTimestamps = append(cv.StatusTimestamps, convertConfigurationVersionStatusTimestamps(ts))
	}
	return cv
}

func convertConfigurationVersionStatusTimestamps(r configurationVersionStatusTimestamp) otf.ConfigurationVersionStatusTimestamp {
	return otf.ConfigurationVersionStatusTimestamp{
		Status:    otf.ConfigurationStatus(*r.GetStatus()),
		Timestamp: r.GetTimestamp(),
	}
}
