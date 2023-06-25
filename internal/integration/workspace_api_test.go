package integration

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/DataDog/jsonapi"
	"github.com/leg100/otf/internal/api/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegration_WorkspaceAPI tests the option to retrieve latest state
// outputs alongside a workspace from the API.
func TestIntegration_WorkspaceAPI_IncludeOutputs(t *testing.T) {
	t.Parallel()

	svc, _, ctx := setup(t, nil)
	sv := svc.createStateVersion(t, ctx, nil)
	_, token := svc.createToken(t, ctx, nil)

	u := fmt.Sprintf("https://%s/api/v2/workspaces/%s?include=outputs", svc.Hostname(), sv.WorkspaceID)
	r, err := http.NewRequest("GET", u, nil)
	require.NoError(t, err)
	r.Header.Add("Authorization", "Bearer "+string(token))

	resp, err := http.DefaultClient.Do(r)
	require.NoError(t, err)
	defer resp.Body.Close()
	if !assert.Equal(t, 200, resp.StatusCode) {
		var buf bytes.Buffer
		io.Copy(&buf, resp.Body)
		t.Log(buf.String())
		return
	}

	got := &types.Workspace{}

	b, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	err = jsonapi.Unmarshal(b, got)
	require.NoError(t, err)

	assert.Equal(t, sv.WorkspaceID, got.ID)
	if assert.NotEmpty(t, got.Outputs) {
		assert.Equal(t, 3, len(got.Outputs))
	}
}
