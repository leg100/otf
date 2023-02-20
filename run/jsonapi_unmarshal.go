package run

import (
	"bytes"

	"github.com/leg100/otf"
	"github.com/leg100/otf/http/jsonapi"
)

func UnmarshalJSONAPI(b []byte) (*Run, error) {
	// unmarshal into json:api struct
	var jrun jsonapi.Run
	if err := jsonapi.UnmarshalPayload(bytes.NewReader(b), &jrun); err != nil {
		return nil, err
	}
	// convert json:api struct to run
	return newFromJSONAPI(&jrun), nil
}

func newFromJSONAPI(from *jsonapi.Run) *Run {
	return &Run{
		id:                     from.ID,
		createdAt:              from.CreatedAt,
		forceCancelAvailableAt: from.ForceCancelAvailableAt,
		isDestroy:              from.IsDestroy,
		executionMode:          otf.ExecutionMode(from.ExecutionMode),
		message:                from.Message,
		positionInQueue:        from.PositionInQueue,
		refresh:                from.Refresh,
		refreshOnly:            from.RefreshOnly,
		status:                 otf.RunStatus(from.Status),
		// TODO: unmarshal timestamps
		replaceAddrs:           from.ReplaceAddrs,
		targetAddrs:            from.TargetAddrs,
		workspaceID:            from.Workspace.ID,
		configurationVersionID: from.ConfigurationVersion.ID,
		// TODO: unmarshal plan and apply relations
	}
}

// newListFromJSONAPI constructs a run list from a json:api struct
func newListFromJSONAPI(from *jsonapi.RunList) *RunList {
	to := RunList{
		Pagination: otf.NewPaginationFromJSONAPI(from.Pagination),
	}
	for _, i := range from.Items {
		to.Items = append(to.Items, newFromJSONAPI(i))
	}
	return &to
}
