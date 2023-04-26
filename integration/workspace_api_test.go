package integration

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/leg100/otf/http/jsonapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegration_WorkspaceAPI(t *testing.T) {
	t.Parallel()

	svc := setup(t, nil)
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

	got := &jsonapi.Workspace{}
	err = jsonapi.UnmarshalPayload(resp.Body, got)
	require.NoError(t, err)

	assert.Equal(t, sv.WorkspaceID, got.ID)
	if assert.NotEmpty(t, got.Outputs) {
		assert.Equal(t, 3, len(got.Outputs))
	}
}
