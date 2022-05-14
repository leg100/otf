package sql

import (
	"time"

	"github.com/leg100/otf"
)

type runResult struct {
	RunID                 *string                 `json:"run_id"`
	PlanID                *string                 `json:"plan_id"`
	ApplyID               *string                 `json:"apply_id"`
	CreatedAt             time.Time               `json:"created_at"`
	UpdatedAt             time.Time               `json:"updated_at"`
	IsDestroy             *bool                   `json:"is_destroy"`
	PositionInQueue       *int32                  `json:"position_in_queue"`
	Refresh               *bool                   `json:"refresh"`
	RefreshOnly           *bool                   `json:"refresh_only"`
	Status                *string                 `json:"status"`
	PlanStatus            *string                 `json:"plan_status"`
	ApplyStatus           *string                 `json:"apply_status"`
	ReplaceAddrs          []string                `json:"replace_addrs"`
	TargetAddrs           []string                `json:"target_addrs"`
	PlanResourceReport    *otf.ResourceReport     `json:"plan_resource_report"`
	ApplyResourceReport   *otf.ResourceReport     `json:"apply_resource_report"`
	ConfigurationVersion  ConfigurationVersions   `json:"configuration_version"`
	Workspace             Workspaces              `json:"workspace"`
	RunStatusTimestamps   []RunStatusTimestamps   `json:"run_status_timestamps"`
	PlanStatusTimestamps  []PlanStatusTimestamps  `json:"plan_status_timestamps"`
	ApplyStatusTimestamps []ApplyStatusTimestamps `json:"apply_status_timestamps"`
}

type runListResult struct {
	RunID                 *string                 `json:"run_id"`
	PlanID                *string                 `json:"plan_id"`
	ApplyID               *string                 `json:"apply_id"`
	CreatedAt             time.Time               `json:"created_at"`
	UpdatedAt             time.Time               `json:"updated_at"`
	IsDestroy             *bool                   `json:"is_destroy"`
	PositionInQueue       *int32                  `json:"position_in_queue"`
	Refresh               *bool                   `json:"refresh"`
	RefreshOnly           *bool                   `json:"refresh_only"`
	Status                *string                 `json:"status"`
	PlanStatus            *string                 `json:"plan_status"`
	ApplyStatus           *string                 `json:"apply_status"`
	ReplaceAddrs          []string                `json:"replace_addrs"`
	TargetAddrs           []string                `json:"target_addrs"`
	PlanResourceReport    *otf.ResourceReport     `json:"plan_resource_report"`
	ApplyResourceReport   *otf.ResourceReport     `json:"apply_resource_report"`
	ConfigurationVersion  ConfigurationVersions   `json:"configuration_version"`
	Workspace             Workspaces              `json:"workspace"`
	RunStatusTimestamps   []RunStatusTimestamps   `json:"run_status_timestamps"`
	PlanStatusTimestamps  []PlanStatusTimestamps  `json:"plan_status_timestamps"`
	ApplyStatusTimestamps []ApplyStatusTimestamps `json:"apply_status_timestamps"`
	FullCount             *int                    `json:"full_count"`
}

func convertRunList(results []runListResult) (items []*otf.Run) {
	for _, r := range results {
		run := otf.Run{
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
				ID:               *r.ApplyID,
				RunID:            *r.RunID,
				Status:           otf.ApplyStatus(*r.ApplyStatus),
				StatusTimestamps: convertApplyStatusTimestamps(r.ApplyStatusTimestamps),
			},
			Plan: &otf.Plan{
				ID:               *r.PlanID,
				RunID:            *r.RunID,
				Status:           otf.PlanStatus(*r.PlanStatus),
				StatusTimestamps: convertPlanStatusTimestamps(r.PlanStatusTimestamps),
			},
			ConfigurationVersion: convertConfigurationVersionComposite(r.ConfigurationVersion),
			Workspace:            convertWorkspaceComposite(r.Workspace),
			StatusTimestamps:     convertRunStatusTimestamps(r.RunStatusTimestamps),
		}
		items = append(items, &run)
	}

	return items
}

func convertRun(r runResult) *otf.Run {
	return &otf.Run{
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
			RunID:            *r.RunID,
			ID:               *r.ApplyID,
			Status:           otf.ApplyStatus(*r.ApplyStatus),
			StatusTimestamps: convertApplyStatusTimestamps(r.ApplyStatusTimestamps),
		},
		Plan: &otf.Plan{
			ID:               *r.PlanID,
			RunID:            *r.RunID,
			Status:           otf.PlanStatus(*r.PlanStatus),
			StatusTimestamps: convertPlanStatusTimestamps(r.PlanStatusTimestamps),
		},
		ConfigurationVersion: convertConfigurationVersionComposite(r.ConfigurationVersion),
		Workspace:            convertWorkspaceComposite(r.Workspace),
		StatusTimestamps:     convertRunStatusTimestamps(r.RunStatusTimestamps),
	}
}

func convertRunStatusTimestamps(rows []RunStatusTimestamps) (timestamps []otf.RunStatusTimestamp) {
	for _, r := range rows {
		timestamps = append(timestamps, otf.RunStatusTimestamp{
			Status:    otf.RunStatus(*r.Status),
			Timestamp: r.Timestamp,
		})
	}
	return timestamps
}

func convertPlanStatusTimestamps(rows []PlanStatusTimestamps) (timestamps []otf.PlanStatusTimestamp) {
	for _, r := range rows {
		timestamps = append(timestamps, otf.PlanStatusTimestamp{
			Status:    otf.PlanStatus(*r.Status),
			Timestamp: r.Timestamp,
		})
	}
	return timestamps
}

func convertApplyStatusTimestamps(rows []ApplyStatusTimestamps) (timestamps []otf.ApplyStatusTimestamp) {
	for _, r := range rows {
		timestamps = append(timestamps, otf.ApplyStatusTimestamp{
			Status:    otf.ApplyStatus(*r.Status),
			Timestamp: r.Timestamp,
		})
	}
	return timestamps
}
