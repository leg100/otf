package integration

import (
	"testing"

	"github.com/chromedp/chromedp"
	"github.com/leg100/otf/internal/run"
	"github.com/stretchr/testify/require"
)

// TestIntegration_StateUI demonstrates the displaying of terraform state via
// the UI
func TestIntegration_StateUI(t *testing.T) {
	integrationTest(t)

	daemon, org, ctx := setup(t, nil)

	// watch run events
	sub, unsub := daemon.WatchRuns()
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
			err := daemon.Apply(ctx, r.ID)
			require.NoError(t, err)
		case run.RunErrored:
			t.Fatal("run unexpectedly finished with an error")
		}
	}

	browser.Run(t, ctx, chromedp.Tasks{
		chromedp.Navigate(workspaceURL(daemon.Hostname(), org.Name, ws.Name)),
		matchRegex(t, `//label[@id='resources-label']`, `Resources \(1\)`),
		matchRegex(t, `//label[@id='outputs-label']`, `Outputs \(0\)`),
		matchText(t, `//table[@id='resources-table']/tbody/tr/td[1]`, `test`),
		matchText(t, `//table[@id='resources-table']/tbody/tr/td[2]`, `hashicorp/null`),
		matchText(t, `//table[@id='resources-table']/tbody/tr/td[3]`, `null_resource`),
		matchText(t, `//table[@id='resources-table']/tbody/tr/td[4]`, `root`),
	})
}
