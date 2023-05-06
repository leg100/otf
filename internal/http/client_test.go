package http

import (
	"bytes"
	"testing"

	"github.com/DataDog/jsonapi"
	"github.com/leg100/otf/api/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_UnmarshalResponse(t *testing.T) {
	want := types.WorkspaceList{
		Items: []*types.Workspace{
			{ID: "ws-1"},
			{ID: "ws-2"},
		},
		Pagination: &types.Pagination{},
	}
	b, err := jsonapi.Marshal(&want.Items, jsonapi.MarshalMeta(want.Pagination))
	require.NoError(t, err)

	var got types.WorkspaceList
	err = unmarshalResponse(bytes.NewReader(b), &got)
	require.NoError(t, err)

	assert.Equal(t, want, got)
}
