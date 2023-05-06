package run

import (
	"github.com/DataDog/jsonapi"
	"github.com/leg100/otf"
	"github.com/leg100/otf/api/types"
	"github.com/leg100/otf/workspace"
)

func UnmarshalJSONAPI(b []byte) (*Run, error) {
	// unmarshal into json:api struct
	var run types.Run
	if err := jsonapi.Unmarshal(b, &run); err != nil {
		return nil, err
	}
	// convert json:api struct to run
	return newFromJSONAPI(&run), nil
}

func newFromJSONAPI(from *types.Run) *Run {
	return &Run{
		ID:                     from.ID,
		CreatedAt:              from.CreatedAt,
		ForceCancelAvailableAt: from.ForceCancelAvailableAt,
		IsDestroy:              from.IsDestroy,
		ExecutionMode:          workspace.ExecutionMode(from.ExecutionMode),
		Message:                from.Message,
		PositionInQueue:        from.PositionInQueue,
		Refresh:                from.Refresh,
		RefreshOnly:            from.RefreshOnly,
		Status:                 otf.RunStatus(from.Status),
		// TODO: unmarshal timestamps
		ReplaceAddrs:           from.ReplaceAddrs,
		TargetAddrs:            from.TargetAddrs,
		WorkspaceID:            from.Workspace.ID,
		ConfigurationVersionID: from.ConfigurationVersion.ID,
		// TODO: unmarshal plan and apply relations
	}
}

// newListFromJSONAPI constructs a run list from a json:api struct
func newListFromJSONAPI(from *types.RunList) *RunList {
	to := RunList{
		Pagination: otf.NewPaginationFromJSONAPI(from.Pagination),
	}
	for _, i := range from.Items {
		to.Items = append(to.Items, newFromJSONAPI(i))
	}
	return &to
}
