package integration

import (
	"testing"

	"github.com/chromedp/chromedp"
	"github.com/stretchr/testify/require"
)

// TestStartRunUI tests starting a run via the Web UI before confirming and
// applying the run.
func TestStartRunUI(t *testing.T) {
	t.Parallel()

	svc := setup(t, nil)
	user, ctx := svc.createUserCtx(t, ctx)
	org := svc.createOrganization(t, ctx)
	root := newRootModule(t, svc.Hostname(), org.Name, "my-test-workspace")

	// in browser, create workspace
	browser := createBrowserCtx(t)
	err := chromedp.Run(browser,
		newSession(t, ctx, svc.Hostname(), user.Username, svc.Secret),
		createWorkspace(t, svc.Hostname(), org.Name, "my-test-workspace"),
	)
	require.NoError(t, err)

	//
	// start run UI functionality requires an existing config version, so
	// create one first by running a plan via the CLI
	//

	// terraform init
	svc.tfcli(t, ctx, "init", root)
	out := svc.tfcli(t, ctx, "plan", root)
	require.Contains(t, out, "Plan: 1 to add, 0 to change, 0 to destroy.")

	// now we have a config version, start a run via the browser
	err = chromedp.Run(browser, startRunTasks(t, svc.Hostname(), org.Name, "my-test-workspace"))
	require.NoError(t, err)
}
