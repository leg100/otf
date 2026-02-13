package integration

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/DataDog/jsonapi"
	tfe "github.com/hashicorp/go-tfe"
	"github.com/leg100/otf/internal/github"
	otfrun "github.com/leg100/otf/internal/run"
	"github.com/leg100/otf/internal/runstatus"
	"github.com/leg100/otf/internal/testutils"
	"github.com/leg100/otf/internal/vcs"
	"github.com/leg100/otf/internal/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegration_WorkspaceAPI tests the option to retrieve latest state
// outputs alongside a workspace from the API.
func TestIntegration_WorkspaceAPI_IncludeOutputs(t *testing.T) {
	integrationTest(t)

	svc, _, ctx := setup(t)
	sv := svc.createStateVersion(t, ctx, nil)
	_, token := svc.createToken(t, ctx, nil)

	u := fmt.Sprintf("https://%s/api/v2/workspaces/%s?include=outputs", svc.System.Hostname(), sv.WorkspaceID)
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

	got := &workspace.TFEWorkspace{}

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
// workspace connected to a VCS repo via the API, and then creating a run that
// sources configuration from the repo.
func TestIntegration_WorkspaceAPI_CreateConnected(t *testing.T) {
	integrationTest(t)

	// setup daemon along with fake github repo
	repo := vcs.NewRandomRepo()
	daemon, org, ctx := setup(t, withGithubOptions(
		github.WithRepo(repo),
		github.WithCommit("0335fb07bb0244b7a169ee89d15c7703e4aaf7de"),
		github.WithArchive(testutils.ReadFile(t, "../testdata/github.tar.gz")),
	))

	_, token := daemon.createToken(t, ctx, nil)

	client, err := tfe.NewClient(&tfe.Config{
		Address: daemon.System.URL("/"),
		Token:   string(token),
	})
	require.NoError(t, err)

	provider := daemon.createVCSProvider(t, ctx, org, nil)

	oauth, err := client.OAuthClients.Create(ctx, org.Name.String(), tfe.OAuthClientCreateOptions{
		OAuthToken:      provider.Token,
		APIURL:          new(daemon.GithubHostname.String()),
		HTTPURL:         new(daemon.GithubHostname.String()),
		ServiceProvider: tfe.ServiceProvider(tfe.ServiceProviderGithub),
	})
	require.NoError(t, err)

	ws, err := client.Workspaces.Create(ctx, org.Name.String(), tfe.WorkspaceCreateOptions{
		Name: new("testing"),
		VCSRepo: &tfe.VCSRepoOptions{
			OAuthTokenID: new(oauth.ID),
			Identifier:   new(repo.String()),
		},
	})
	require.NoError(t, err)

	run, err := daemon.Runs.Create(ctx, testutils.ParseID(t, ws.ID), otfrun.CreateOptions{})
	require.NoError(t, err)

	run = daemon.waitRunStatus(t, ctx, run.ID, runstatus.Planned)
	// status matches, now check whether reports match as well
	assert.Equal(t, &otfrun.Report{Additions: 2}, run.Plan.ResourceReport)
}
