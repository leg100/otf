package sql

import (
	"time"

	"github.com/leg100/otf"
)

type runResult struct {
	RunID                       *string                 `json:"run_id"`
	CreatedAt                   time.Time               `json:"created_at"`
	UpdatedAt                   time.Time               `json:"updated_at"`
	IsDestroy                   *bool                   `json:"is_destroy"`
	PositionInQueue             *int32                  `json:"position_in_queue"`
	Refresh                     *bool                   `json:"refresh"`
	RefreshOnly                 *bool                   `json:"refresh_only"`
	Status                      *string                 `json:"status"`
	PlanStatus                  *string                 `json:"plan_status"`
	PlannedResourceAdditions    *int32                  `json:"planned_resource_additions"`
	PlannedResourceChanges      *int32                  `json:"planned_resource_changes"`
	PlannedResourceDestructions *int32                  `json:"planned_resource_destructions"`
	ApplyStatus                 *string                 `json:"apply_status"`
	AppliedResourceAdditions    *int32                  `json:"applied_resource_additions"`
	AppliedResourceChanges      *int32                  `json:"applied_resource_changes"`
	AppliedResourceDestructions *int32                  `json:"applied_resource_destructions"`
	ReplaceAddrs                []string                `json:"replace_addrs"`
	TargetAddrs                 []string                `json:"target_addrs"`
	ConfigurationVersion        ConfigurationVersions   `json:"configuration_version"`
	Workspace                   Workspaces              `json:"workspace"`
	RunStatusTimestamps         []RunStatusTimestamps   `json:"run_status_timestamps"`
	PlanStatusTimestamps        []PlanStatusTimestamps  `json:"plan_status_timestamps"`
	ApplyStatusTimestamps       []ApplyStatusTimestamps `json:"apply_status_timestamps"`
}

type runListResult struct {
	RunID                       *string                 `json:"run_id"`
	CreatedAt                   time.Time               `json:"created_at"`
	UpdatedAt                   time.Time               `json:"updated_at"`
	IsDestroy                   *bool                   `json:"is_destroy"`
	PositionInQueue             *int32                  `json:"position_in_queue"`
	Refresh                     *bool                   `json:"refresh"`
	RefreshOnly                 *bool                   `json:"refresh_only"`
	Status                      *string                 `json:"status"`
	PlanStatus                  *string                 `json:"plan_status"`
	PlannedResourceAdditions    *int32                  `json:"planned_resource_additions"`
	PlannedResourceChanges      *int32                  `json:"planned_resource_changes"`
	PlannedResourceDestructions *int32                  `json:"planned_resource_destructions"`
	ApplyStatus                 *string                 `json:"apply_status"`
	AppliedResourceAdditions    *int32                  `json:"applied_resource_additions"`
	AppliedResourceChanges      *int32                  `json:"applied_resource_changes"`
	AppliedResourceDestructions *int32                  `json:"applied_resource_destructions"`
	ReplaceAddrs                []string                `json:"replace_addrs"`
	TargetAddrs                 []string                `json:"target_addrs"`
	ConfigurationVersion        ConfigurationVersions   `json:"configuration_version"`
	Workspace                   Workspaces              `json:"workspace"`
	RunStatusTimestamps         []RunStatusTimestamps   `json:"run_status_timestamps"`
	PlanStatusTimestamps        []PlanStatusTimestamps  `json:"plan_status_timestamps"`
	ApplyStatusTimestamps       []ApplyStatusTimestamps `json:"apply_status_timestamps"`
	FullCount                   *int                    `json:"full_count"`
}

func convertRunList(results []runListResult) (items []*otf.Run) {
	for _, r := range results {
		items = append(items, &otf.Run{
			ID: *r.RunID,
			Timestamps: otf.Timestamps{
				CreatedAt: r.CreatedAt,
				UpdatedAt: r.UpdatedAt,
			},
			IsDestroy:       *r.IsDestroy,
			PositionInQueue: int(*r.PositionInQueue),
			Refresh:         *r.Refresh,
			RefreshOnly:     *r.RefreshOnly,
			Status:          otf.RunStatus(*r.Status),
			ReplaceAddrs:    r.ReplaceAddrs,
			TargetAddrs:     r.TargetAddrs,
			Apply: &otf.Apply{
				Status:           otf.ApplyStatus(*r.ApplyStatus),
				StatusTimestamps: convertApplyStatusTimestamps(r.ApplyStatusTimestamps),
				Resources: otf.Resources{
					ResourceAdditions: int(*r.AppliedResourceAdditions),
				},
			},
			Plan: &otf.Plan{
				Status:           otf.PlanStatus(*r.PlanStatus),
				StatusTimestamps: convertPlanStatusTimestamps(r.PlanStatusTimestamps),
				Resources: otf.Resources{
					ResourceAdditions: int(*r.PlannedResourceAdditions),
				},
			},
			ConfigurationVersion: convertConfigurationVersionComposite(r.ConfigurationVersion),
			Workspace:            convertWorkspaceComposite(r.Workspace),
			StatusTimestamps:     convertRunStatusTimestamps(r.RunStatusTimestamps),
		})
	}

	return items
}
func convertRun(result runResult) *otf.Run {
	return &otf.Run{
		ID: *result.RunID,
		Timestamps: otf.Timestamps{
			CreatedAt: result.CreatedAt,
			UpdatedAt: result.UpdatedAt,
		},
		IsDestroy:       *result.IsDestroy,
		PositionInQueue: int(*result.PositionInQueue),
		Refresh:         *result.Refresh,
		RefreshOnly:     *result.RefreshOnly,
		Status:          otf.RunStatus(*result.Status),
		ReplaceAddrs:    result.ReplaceAddrs,
		TargetAddrs:     result.TargetAddrs,
		Apply: &otf.Apply{
			Status:           otf.ApplyStatus(*result.ApplyStatus),
			StatusTimestamps: convertApplyStatusTimestamps(result.ApplyStatusTimestamps),
			Resources: otf.Resources{
				ResourceAdditions: int(*result.AppliedResourceAdditions),
			},
		},
		Plan: &otf.Plan{
			Status:           otf.PlanStatus(*result.PlanStatus),
			StatusTimestamps: convertPlanStatusTimestamps(result.PlanStatusTimestamps),
			Resources: otf.Resources{
				ResourceAdditions: int(*result.PlannedResourceAdditions),
			},
		},
		ConfigurationVersion: convertConfigurationVersionComposite(result.ConfigurationVersion),
		Workspace:            convertWorkspaceComposite(result.Workspace),
		StatusTimestamps:     convertRunStatusTimestamps(result.RunStatusTimestamps),
	}
}

func convertRunStatusTimestamps(rows []RunStatusTimestamps) (timestamps []otf.RunStatusTimestamp) {
	for _, r := range rows {
		timestamps = append(timestamps, otf.RunStatusTimestamp{
			Status:    otf.RunStatus(*r.GetStatus()),
			Timestamp: r.GetTimestamp(),
		})
	}
	return timestamps
}

func convertPlanStatusTimestamps(rows []PlanStatusTimestamps) (timestamps []otf.PlanStatusTimestamp) {
	for _, r := range rows {
		timestamps = append(timestamps, otf.PlanStatusTimestamp{
			Status:    otf.PlanStatus(*r.GetStatus()),
			Timestamp: r.GetTimestamp(),
		})
	}
	return timestamps
}

func convertApplyStatusTimestamps(rows []ApplyStatusTimestamps) (timestamps []otf.ApplyStatusTimestamp) {
	for _, r := range rows {
		timestamps = append(timestamps, otf.ApplyStatusTimestamp{
			Status:    otf.ApplyStatus(*r.GetStatus()),
			Timestamp: r.GetTimestamp(),
		})
	}
	return timestamps
}
