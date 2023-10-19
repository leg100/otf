package integration

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"testing"
	"time"

	expect "github.com/google/goexpect"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/run"
	"github.com/leg100/otf/internal/testutils"
	"github.com/stretchr/testify/require"
)

// TestIntegration_TerraformCLICancel demonstrates a user canceling a run via
// the terraform CLI.
func TestIntegration_TerraformCLICancel(t *testing.T) {
	integrationTest(t)

	svc, org, ctx := setup(t, nil)

	// Canceling a run is not straight-forward, because to do so reliably the
	// terraform apply should be interrupted precisely when it is in mid-flow,
	// i.e. while it is planning. To achieve this, the test uses the 'http'
	// data source, which contacts a test handler that never returns a response
	// and so should cause terraform plan to hang. At this point the interrupt
	// can be sent.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// never return
		<-make(chan struct{})
	}))

	// create some config and run terraform init
	config := newRootModule(t, svc.Hostname(), org.Name, t.Name(), fmt.Sprintf(`
data "http" "wait" {
	url = "%s"
}
`, srv.URL))
	svc.tfcli(t, ctx, "init", config)

	tfpath := svc.downloadTerraform(t, ctx, nil)

	out, err := os.CreateTemp(t.TempDir(), "terraform-cli-cancel.out")
	require.NoError(t, err)

	// Invoke terraform apply
	_, token := svc.createToken(t, ctx, nil)
	e, tferr, err := expect.SpawnWithArgs(
		[]string{tfpath, "-chdir=" + config, "apply", "-no-color"},
		time.Minute,
		expect.PartialMatch(true),
		expect.Tee(out),
		expect.SetEnv(
			append(sharedEnvs, internal.CredentialEnv(svc.Hostname(), token)),
		),
	)
	require.NoError(t, err)
	defer e.Close()

	// Wait for apply to start reading http data source that never returns
	_, _, err = e.Expect(regexp.MustCompile(`data\.http\.wait: Reading\.\.\.`), time.Second*10)
	require.NoError(t, err)

	// Send Ctrl-C now that terraform apply is in-flow.
	e.SendSignal(os.Interrupt)

	// Confirm canceling run
	e.ExpectBatch([]expect.Batcher{
		&expect.BExp{R: "Do you want to cancel the remote operation?"},
		&expect.BExp{R: "Enter a value:"}, &expect.BSnd{S: "yes\n"},
		&expect.BExp{R: "The remote operation was successfully cancelled."},
	}, time.Minute)
	// Terraform should return with exit code 0
	require.NoError(t, <-tferr, string(testutils.ReadFile(t, out.Name())))
	t.Log(string(testutils.ReadFile(t, out.Name())))

	runs, err := svc.ListRuns(ctx, run.ListOptions{Organization: &org.Name})
	require.NoError(t, err)
	require.Equal(t, 1, len(runs.Items))
	require.Equal(t, internal.RunCanceled, runs.Items[0].Status)
}
