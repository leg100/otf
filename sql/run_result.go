package sql

import (
	"time"

	"github.com/leg100/otf"
)

type runResult interface {
	runResultWithoutRelations

	GetRunStatusTimestamps() []RunStatusTimestamps
	GetPlan() Plans
	GetApply() Applies
	GetWorkspace() Workspaces
	GetConfigurationVersion() ConfigurationVersions
}

type runResultWithoutRelations interface {
	GetRunID() *string
	GetIsDestroy() *bool
	GetWorkspaceID() *string
	GetStatus() *string
	GetConfigurationVersionID() *string
	GetReplaceAddrs() []string
	GetTargetAddrs() []string

	Timestamps
}

type runListResult interface {
	runResult

	GetFullCount() *int
}

type runStatusTimestamp interface {
	GetRunID() *string
	GetStatus() *string
	GetTimestamp() time.Time
}

func addResultToRun(run *otf.Run, result runResultWithoutRelations) {
	run.ID = *result.GetRunID()
	run.Timestamps = convertTimestamps(result)
	run.Status = otf.RunStatus(*result.GetStatus())
	run.IsDestroy = *result.GetIsDestroy()
	run.ReplaceAddrs = result.GetReplaceAddrs()
	run.TargetAddrs = result.GetTargetAddrs()
}

func convertRunWithoutRelations(result runResultWithoutRelations) *otf.Run {
	var run otf.Run
	addResultToRun(&run, result)
	return &run
}

func convertRun(result runResult) *otf.Run {
	run := convertRunWithoutRelations(result)
	run.Plan = convertPlan(result.GetPlan())
	run.Apply = convertApply(result.GetApply())
	run.Workspace = convertWorkspaceComposite(result.GetWorkspace())
	run.ConfigurationVersion = convertConfigurationVersionComposite(result.GetConfigurationVersion())

	for _, ts := range result.GetRunStatusTimestamps() {
		run.StatusTimestamps = append(run.StatusTimestamps, convertRunStatusTimestamp(ts))
	}

	return run
}

func convertRunStatusTimestamp(r runStatusTimestamp) otf.RunStatusTimestamp {
	return otf.RunStatusTimestamp{
		Status:    otf.RunStatus(*r.GetStatus()),
		Timestamp: r.GetTimestamp(),
	}
}
