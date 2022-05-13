package sql

import (
	"time"

	"github.com/leg100/otf"
)

type runListResult struct {
	RunID                  *string               `json:"run_id"`
	CreatedAt              time.Time             `json:"created_at"`
	UpdatedAt              time.Time             `json:"updated_at"`
	IsDestroy              *bool                 `json:"is_destroy"`
	PositionInQueue        *int32                `json:"position_in_queue"`
	Refresh                *bool                 `json:"refresh"`
	RefreshOnly            *bool                 `json:"refresh_only"`
	Status                 *string               `json:"status"`
	ReplaceAddrs           []string              `json:"replace_addrs"`
	TargetAddrs            []string              `json:"target_addrs"`
	WorkspaceID            *string               `json:"workspace_id"`
	ConfigurationVersionID *string               `json:"configuration_version_id"`
	Plan                   Plans                 `json:"plan"`
	Apply                  Applies               `json:"apply"`
	ConfigurationVersion   ConfigurationVersions `json:"configuration_version"`
	Workspace              Workspaces            `json:"workspace"`
	RunStatusTimestamps    []RunStatusTimestamps `json:"run_status_timestamps"`
	FullCount              *int                  `json:"full_count"`
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
	run.ConfigurationVersion = convertConfigurationVersionResult(result.GetConfigurationVersion())

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
