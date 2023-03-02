package run

import (
	"bytes"

	"github.com/leg100/otf"
	"github.com/leg100/otf/http/jsonapi"
)

func UnmarshalJSONAPI(b []byte) (*otf.Run, error) {
	// unmarshal into json:api struct
	var jrun jsonapi.Run
	if err := jsonapi.UnmarshalPayload(bytes.NewReader(b), &jrun); err != nil {
		return nil, err
	}
	// convert json:api struct to run
	return newFromJSONAPI(&jrun), nil
}

func newFromJSONAPI(from *jsonapi.Run) *otf.Run {
	return &otf.Run{
		ID:                     from.ID,
		CreatedAt:              from.CreatedAt,
		ForceCancelAvailableAt: from.ForceCancelAvailableAt,
		IsDestroy:              from.IsDestroy,
		ExecutionMode:          otf.ExecutionMode(from.ExecutionMode),
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
func newListFromJSONAPI(from *jsonapi.RunList) *otf.RunList {
	to := otf.RunList{
		Pagination: otf.NewPaginationFromJSONAPI(from.Pagination),
	}
	for _, i := range from.Items {
		to.Items = append(to.Items, newFromJSONAPI(i))
	}
	return &to
}
