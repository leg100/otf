package run

import (
	"bytes"

	"github.com/leg100/otf"
	"github.com/leg100/otf/http/jsonapi"
	"github.com/leg100/otf/workspace"
)

func UnmarshalJSONAPI(b []byte) (*Run, error) {
	// unmarshal into json:api struct
	var run jsonapi.Run
	if err := jsonapi.UnmarshalPayload(bytes.NewReader(b), &run); err != nil {
		return nil, err
	}
	// convert json:api struct to run
	return newFromJSONAPI(&run), nil
}

func newFromJSONAPI(from *jsonapi.Run) *Run {
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
func newListFromJSONAPI(from *jsonapi.RunList) *RunList {
	to := RunList{
		Pagination: jsonapi.NewPaginationFromJSONAPI(from.Pagination),
	}
	for _, i := range from.Items {
		to.Items = append(to.Items, newFromJSONAPI(i))
	}
	return &to
}
