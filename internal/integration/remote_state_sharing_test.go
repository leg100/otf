package integration

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/runstatus"
	"github.com/leg100/otf/internal/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRemoteStateSharing demonstrates the use of terraform_remote_state, and
// permitting or denying its use.
func TestRemoteStateSharing(t *testing.T) {
	integrationTest(t)

	daemon, org, ctx := setup(t)
	// producer is the workspace sharing its state
	producer, err := daemon.Workspaces.Create(ctx, workspace.CreateOptions{
		Name:              new("producer"),
		Organization:      &org.Name,
		GlobalRemoteState: new(true),
	})
	require.NoError(t, err)
	// consumer is the workspace consuming the state of the producer
	consumer := daemon.createWorkspace(t, ctx, org)

	// populate producer with state
	producerRoot := t.TempDir()
	producerConfig := `output "foo" { value = "bar" }`
	err = os.WriteFile(filepath.Join(producerRoot, "main.tf"), []byte(producerConfig), 0o777)
	require.NoError(t, err)
	tarball, err := internal.Pack(producerRoot)
	require.NoError(t, err)
	producerCV := daemon.createConfigurationVersion(t, ctx, producer, nil)
	err = daemon.Configs.UploadConfig(ctx, producerCV.ID, tarball)
	require.NoError(t, err)

	producerRun := daemon.createRun(t, ctx, producer, producerCV, nil)

	// Wait for run to reach planned state before applying
	planned := daemon.waitRunStatus(t, ctx, producerRun.ID, runstatus.Planned)
	err = daemon.Runs.Apply(ctx, planned.ID)
	require.NoError(t, err)

	// Wait for run to be applied
	daemon.waitRunStatus(t, ctx, producerRun.ID, runstatus.Applied)

	// consume state in a run in the consumer workspace from the producer
	// workspace.
	consumerRoot := t.TempDir()
	consumerConfig := fmt.Sprintf(`
data "terraform_remote_state" "producer" {
  backend = "remote"

  config = {
	hostname = "%s"
    organization = "%s"
    workspaces = {
      name = "%s"
    }
  }
}

output "remote_foo" {
  value = data.terraform_remote_state.producer.outputs.foo
}
`, daemon.System.Hostname(), org.Name, producer.Name)
	err = os.WriteFile(filepath.Join(consumerRoot, "main.tf"), []byte(consumerConfig), 0o777)
	require.NoError(t, err)
	tarball, err = internal.Pack(consumerRoot)
	require.NoError(t, err)
	consumerCV := daemon.createConfigurationVersion(t, ctx, consumer, nil)
	err = daemon.Configs.UploadConfig(ctx, consumerCV.ID, tarball)
	require.NoError(t, err)

	// create run and apply
	consumerRun := daemon.createRun(t, ctx, consumer, consumerCV, nil)
	planned = daemon.waitRunStatus(t, ctx, consumerRun.ID, runstatus.Planned)
	err = daemon.Runs.Apply(ctx, planned.ID)
	require.NoError(t, err)
	daemon.waitRunStatus(t, ctx, consumerRun.ID, runstatus.Applied)

	got := daemon.getCurrentState(t, ctx, consumer.ID)
	if assert.Contains(t, got.Outputs, "remote_foo", got.Outputs) {
		assert.Equal(t, `"bar"`, string(got.Outputs["remote_foo"].Value))
	}
}
