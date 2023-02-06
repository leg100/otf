package run

import (
	"github.com/jackc/pgtype"
	"github.com/leg100/otf/http/jsonapi"
	"github.com/leg100/otf/sql/pggen"
)

func UnmarshalRunJSONAPI(d *jsonapi.Run) *Run {
	return &Run{
		id:                     d.ID,
		createdAt:              d.CreatedAt,
		forceCancelAvailableAt: d.ForceCancelAvailableAt,
		isDestroy:              d.IsDestroy,
		executionMode:          ExecutionMode(d.ExecutionMode),
		message:                d.Message,
		positionInQueue:        d.PositionInQueue,
		refresh:                d.Refresh,
		refreshOnly:            d.RefreshOnly,
		status:                 RunStatus(d.Status),
		// TODO: unmarshal timestamps
		replaceAddrs:           d.ReplaceAddrs,
		targetAddrs:            d.TargetAddrs,
		workspaceID:            d.Workspace.ID,
		configurationVersionID: d.ConfigurationVersion.ID,
		// TODO: unmarshal plan and apply relations
	}
}

// UnmarshalRunListJSONAPI converts a DTO into a run list
func UnmarshalRunListJSONAPI(json *jsonapi.RunList) *RunList {
	wl := RunList{
		Pagination: UnmarshalPaginationJSONAPI(json.Pagination),
	}
	for _, i := range json.Items {
		wl.Items = append(wl.Items, UnmarshalRunJSONAPI(i))
	}

	return &wl
}

func unmarshalRunStatusTimestampRows(rows []pggen.RunStatusTimestamps) (timestamps []RunStatusTimestamp) {
	for _, ty := range rows {
		timestamps = append(timestamps, RunStatusTimestamp{
			Status:    RunStatus(ty.Status.String),
			Timestamp: ty.Timestamp.Time.UTC(),
		})
	}
	return timestamps
}

func unmarshalPlanStatusTimestampRows(rows []pggen.PhaseStatusTimestamps) (timestamps []PhaseStatusTimestamp) {
	for _, ty := range rows {
		timestamps = append(timestamps, PhaseStatusTimestamp{
			Status:    PhaseStatus(ty.Status.String),
			Timestamp: ty.Timestamp.Time.UTC(),
		})
	}
	return timestamps
}

func unmarshalApplyStatusTimestampRows(rows []pggen.PhaseStatusTimestamps) (timestamps []PhaseStatusTimestamp) {
	for _, ty := range rows {
		timestamps = append(timestamps, PhaseStatusTimestamp{
			Status:    PhaseStatus(ty.Status.String),
			Timestamp: ty.Timestamp.Time.UTC(),
		})
	}
	return timestamps
}
