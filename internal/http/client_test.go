package http

import (
	"bytes"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/DataDog/jsonapi"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/tfeapi/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_UnmarshalResponse(t *testing.T) {
	type (
		WorkspaceOutput struct {
			ID        resource.TfeID `jsonapi:"primary,workspace-outputs"`
			Name      string         `jsonapi:"attribute" json:"name"`
			Sensitive bool           `jsonapi:"attribute" json:"sensitive"`
			Type      string         `jsonapi:"attribute" json:"output-type"`
			Value     any            `jsonapi:"attribute" json:"value"`
		}
		Workspace struct {
			ID      resource.TfeID     `jsonapi:"primary,workspaces"`
			Outputs []*WorkspaceOutput `jsonapi:"relationship" json:"outputs"`
		}
		WorkspaceList struct {
			*types.Pagination
			Items []*Workspace
		}
	)
	want := WorkspaceList{
		Items: []*Workspace{
			{ID: resource.NewTfeID(resource.WorkspaceKind), Outputs: []*WorkspaceOutput{}},
			{ID: resource.NewTfeID(resource.WorkspaceKind), Outputs: []*WorkspaceOutput{}},
		},
		Pagination: &types.Pagination{},
	}
	b, err := jsonapi.Marshal(&want.Items, jsonapi.MarshalMeta(want.Pagination))
	require.NoError(t, err)

	var got WorkspaceList
	err = unmarshalResponse(bytes.NewReader(b), &got)
	require.NoError(t, err)

	assert.Equal(t, want, got)
}

func TestClient_checkResponseCode(t *testing.T) {
	tests := []struct {
		name     string
		response *http.Response
		want     error
	}{
		{"200 OK", &http.Response{StatusCode: 200}, nil},
		{"204 No Content", &http.Response{StatusCode: 204}, nil},
		{"401 Not Authorized", &http.Response{StatusCode: 401}, internal.ErrUnauthorized},
		{"404 Not Found", &http.Response{StatusCode: 404}, internal.ErrResourceNotFound},
		{
			"500 Error",
			&http.Response{
				Status: "500 Internal Server Error",
				Body:   newBody(`{"errors":[{"status":"500","title":"Internal Server Error","detail":"cannot marshal unknown type: *types.AgentToken"}]}`),
			},
			errors.New("Internal Server Error: cannot marshal unknown type: *types.AgentToken"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, checkResponseCode(tt.response))
		})
	}
}

type bodyReader struct {
	*strings.Reader
}

func newBody(body string) *bodyReader {
	return &bodyReader{Reader: strings.NewReader(body)}
}

func (r *bodyReader) Close() error { return nil }
