package integration

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/run"
	"github.com/leg100/otf/internal/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRemoteStateSharing demonstrates the use of terraform_remote_state, and
// permitting or denying its use.
func TestRemoteStateSharing(t *testing.T) {
	integrationTest(t)

	daemon, org, ctx := setup(t, nil)
	// producer is the workspace sharing its state
	producer, err := daemon.CreateWorkspace(ctx, workspace.CreateOptions{
		Name:              internal.String("producer"),
		Organization:      internal.String(org.Name),
		GlobalRemoteState: internal.Bool(true),
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
	err = daemon.UploadConfig(ctx, producerCV.ID, tarball)
	require.NoError(t, err)
	// create run and apply
	_ = daemon.createRun(t, ctx, producer, producerCV)
applied:
	for event := range daemon.sub {
		if r, ok := event.Payload.(*run.Run); ok {
			switch r.Status {
			case internal.RunPlanned:
				err := daemon.Apply(ctx, r.ID)
				require.NoError(t, err)
			case internal.RunApplied:
				break applied
			case internal.RunErrored:
				t.Fatalf("run unexpectedly errored")
			}
		}
	}

	// consume state from a run in the consumer workspace
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
`, daemon.Hostname(), org.Name, producer.Name)
	err = os.WriteFile(filepath.Join(consumerRoot, "main.tf"), []byte(consumerConfig), 0o777)
	require.NoError(t, err)
	tarball, err = internal.Pack(consumerRoot)
	require.NoError(t, err)
	consumerCV := daemon.createConfigurationVersion(t, ctx, consumer, nil)
	err = daemon.UploadConfig(ctx, consumerCV.ID, tarball)
	require.NoError(t, err)

	// create run and apply
	_ = daemon.createRun(t, ctx, consumer, consumerCV)
	for event := range daemon.sub {
		if r, ok := event.Payload.(*run.Run); ok {
			switch r.Status {
			case internal.RunPlanned:
				err := daemon.Apply(ctx, r.ID)
				require.NoError(t, err)
			case internal.RunApplied:
				return
			case internal.RunErrored:
				t.Fatalf("run unexpectedly errored")
			}
		}
	}

	got := daemon.getCurrentState(t, ctx, consumer.ID)
	if assert.Contains(t, got.Outputs, "foo") {
		assert.Equal(t, "bar", got.Outputs["foo"])
	}
}
