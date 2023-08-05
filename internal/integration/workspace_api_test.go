package integration

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/DataDog/jsonapi"
	tfe "github.com/hashicorp/go-tfe"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/api"
	"github.com/leg100/otf/internal/api/types"
	"github.com/leg100/otf/internal/cloud"
	"github.com/leg100/otf/internal/github"
	"github.com/leg100/otf/internal/run"
	"github.com/leg100/otf/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegration_WorkspaceAPI tests the option to retrieve latest state
// outputs alongside a workspace from the API.
func TestIntegration_WorkspaceAPI_IncludeOutputs(t *testing.T) {
	integrationTest(t)

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

// TestIntegration_WorkspaceAPI_CreateConnected demonstrates creating a
// worskspace connected to a VCS repo via the API, and then creating a run that
// sources configuration from the repo.
func TestIntegration_WorkspaceAPI_CreateConnected(t *testing.T) {
	integrationTest(t)

	// setup daemon along with fake github repo
	repo := cloud.NewTestRepo()
	daemon, org, ctx := setup(t, nil,
		github.WithRepo(repo),
		github.WithCommit("0335fb07bb0244b7a169ee89d15c7703e4aaf7de"),
		github.WithArchive(testutils.ReadFile(t, "../testdata/github.tar.gz")),
	)

	_, token := daemon.createToken(t, ctx, nil)

	client, err := tfe.NewClient(&tfe.Config{
		Address: "https://" + daemon.Hostname(),
		Token:   string(token),
	})
	require.NoError(t, err)

	provider := daemon.createVCSProvider(t, ctx, org)

	oauth, err := client.OAuthClients.Create(ctx, org.Name, tfe.OAuthClientCreateOptions{
		OAuthToken:      internal.String(provider.Token),
		APIURL:          internal.String(api.GithubAPIURL),
		HTTPURL:         internal.String(api.GithubHTTPURL),
		ServiceProvider: tfe.ServiceProvider(tfe.ServiceProviderGithub),
	})
	require.NoError(t, err)

	ws, err := client.Workspaces.Create(ctx, org.Name, tfe.WorkspaceCreateOptions{
		Name: internal.String("testing"),
		VCSRepo: &tfe.VCSRepoOptions{
			OAuthTokenID: internal.String(oauth.ID),
			Identifier:   internal.String(repo),
		},
	})
	require.NoError(t, err)

	_, err = daemon.CreateRun(ctx, ws.ID, run.CreateOptions{})
	require.NoError(t, err)

	for event := range daemon.sub {
		if r, ok := event.Payload.(*run.Run); ok {
			if r.Status == internal.RunPlanned {
				// status matches, now check whether reports match as well
				assert.Equal(t, &run.Report{Additions: 2}, r.Plan.ResourceReport)
				break
			}
			require.False(t, r.Done(), "run unexpectedly finished with status %s", r.Status)
		}
	}
}
