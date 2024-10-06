package integration

import (
	"regexp"
	"testing"

	"github.com/leg100/otf/internal/run"
	"github.com/stretchr/testify/require"
)

// TestIntegration_StateUI demonstrates the displaying of terraform state via
// the UI
func TestIntegration_StateUI(t *testing.T) {
	integrationTest(t)

	daemon, org, ctx := setup(t, nil)

	// watch run events
	sub, unsub := daemon.Runs.Watch(ctx)
	defer unsub()

	// create run and wait for it to complete
	ws := daemon.createWorkspace(t, ctx, org)
	cv := daemon.createAndUploadConfigurationVersion(t, ctx, ws, nil)
	_ = daemon.createRun(t, ctx, ws, cv)
applied:
	for event := range sub {
		r := event.Payload
		switch r.Status {
		case run.RunApplied:
			break applied
		case run.RunPlanned:
			err := daemon.Runs.Apply(ctx, r.ID)
			require.NoError(t, err)
		case run.RunErrored:
			t.Fatal("run unexpectedly finished with an error")
		}
	}

	page := browser.New(t, ctx)

	_, err := page.Goto(workspaceURL(daemon.System.Hostname(), org.Name, ws.Name))
	require.NoError(t, err)

	err = expect.Locator(page.Locator(`//label[@id='resources-label']`)).ToHaveText(regexp.MustCompile(`Resources \(1\)`))
	require.NoError(t, err)

	err = expect.Locator(page.Locator(`//label[@id='outputs-label']`)).ToHaveText(regexp.MustCompile(`Outputs \(0\)`))
	require.NoError(t, err)

	err = expect.Locator(page.Locator(`//table[@id='resources-table']/tbody/tr/td[1]`)).ToHaveText(`test`)
	require.NoError(t, err)

	err = expect.Locator(page.Locator(`//table[@id='resources-table']/tbody/tr/td[2]`)).ToHaveText(`hashicorp/null`)
	require.NoError(t, err)

	err = expect.Locator(page.Locator(`//table[@id='resources-table']/tbody/tr/td[3]`)).ToHaveText(`null_resource`)
	require.NoError(t, err)

	err = expect.Locator(page.Locator(`//table[@id='resources-table']/tbody/tr/td[4]`)).ToHaveText(`root`)
	require.NoError(t, err)
}
