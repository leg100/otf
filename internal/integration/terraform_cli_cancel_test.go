package integration

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	expect "github.com/google/goexpect"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/run"
	"github.com/stretchr/testify/require"
)

// TestIntegration_TerraformCLICancel demonstrates a user canceling a run via
// the terraform CLI.
func TestIntegration_TerraformCLICancel(t *testing.T) {
	t.Parallel()

	svc := setup(t, nil)
	user, ctx := svc.createUserCtx(t, ctx)
	org := svc.createOrganization(t, ctx)

	// Canceling a run is not straight-forward, because to do so reliably the
	// terraform plan should be interrupted precisely when it is in mid-flow,
	// i.e. while it is planning. To achieve this, the test uses the 'http'
	// data source, which contacts a test handler, which uses channels to
	// co-ordinate sending the interrupt.
	planning := make(chan struct{})
	interrupted := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(planning)
		<-interrupted
	}))

	// create some config and run terraform init
	config := newRootModule(t, svc.Hostname(), org.Name, t.Name(), fmt.Sprintf(`
data "http" "wait" {
	url = "%s"
}
`, srv.URL))
	svc.tfcli(t, ctx, "init", config)

	// Invoke terraform plan
	_, token := svc.createToken(t, ctx, user)
	e, tferr, err := expect.SpawnWithArgs(
		[]string{"terraform", "-chdir=" + config, "plan", "-no-color"},
		time.Minute,
		expect.PartialMatch(true),
		expect.SetEnv(
			append(envs, internal.CredentialEnv(svc.Hostname(), token)),
		),
	)
	require.NoError(t, err)
	defer e.Close()

	// wait for terraform plan to call handler
	<-planning
	// send Ctrl-C now that terraform plan is in-flow.
	e.SendSignal(os.Interrupt)
	close(interrupted)

	// Confirm canceling run
	e.ExpectBatch([]expect.Batcher{
		&expect.BExp{R: "Do you want to cancel the remote operation?"},
		&expect.BExp{R: "Enter a value:"}, &expect.BSnd{S: "yes\n"},
		&expect.BExp{R: "The remote operation was successfully cancelled."},
	}, time.Minute)
	// Terraform should return with exit code 0
	require.NoError(t, <-tferr)

	runs, err := svc.ListRuns(ctx, run.RunListOptions{Organization: &org.Name})
	require.NoError(t, err)
	require.Equal(t, 1, len(runs.Items))
	require.Equal(t, internal.RunCanceled, runs.Items[0].Status)
}
