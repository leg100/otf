package integration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/run"
	"github.com/leg100/otf/internal/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegration_RunStatus creates successive runs on a workspace, each time
// making changes to the configuration, and checking that the change is
// reflected in the run status and resource/output change reports.
func TestIntegration_RunStatus(t *testing.T) {
	integrationTest(t)

	// Create a workspace with auto-apply enabled
	daemon, org, ctx := setup(t, nil)
	ws, err := daemon.CreateWorkspace(ctx, workspace.CreateOptions{
		Name:         internal.String(t.Name()),
		Organization: internal.String(org.Name),
		AutoApply:    internal.Bool(true),
	})
	require.NoError(t, err)

	// watch run events
	sub, unsub := daemon.WatchRuns()
	defer unsub()

	// directory for root module
	root := t.TempDir()

	// create a sequence of steps that the test will execute sequentially.
	steps := []struct {
		name               string
		config             string
		wantStatus         run.Status
		wantResourceReport run.Report
		wantOutputReport   run.Report
	}{
		{
			name:       "add resource",
			config:     `resource "random_pet" "cat" { prefix = "mr-" }`,
			wantStatus: run.RunApplied,
			wantResourceReport: run.Report{
				Additions: 1,
			},
		},
		{
			name: "replace resource",
			config: `resource "random_pet" "cat" { prefix = "sir-" }
`,
			wantStatus: run.RunApplied,
			wantResourceReport: run.Report{
				Additions:    1,
				Destructions: 1,
			},
		},
		{
			name: "new output",
			config: `resource "random_pet" "cat" { prefix = "sir-" }
output "cat_name" { value = random_pet.cat.id }
`,
			wantStatus: run.RunApplied,
			wantOutputReport: run.Report{
				Additions: 1,
			},
		},
		{
			name:       "destroy all",
			wantStatus: run.RunApplied,
			wantResourceReport: run.Report{
				Destructions: 1,
			},
			wantOutputReport: run.Report{
				Destructions: 1,
			},
		},
	}
	for _, step := range steps {
		t.Run(step.name, func(t *testing.T) {
			cv := daemon.createConfigurationVersion(t, ctx, ws, nil)

			// create tarball of root module and upload
			err := os.WriteFile(filepath.Join(root, "main.tf"), []byte(step.config), 0o777)
			require.NoError(t, err)
			tarball, err := internal.Pack(root)
			require.NoError(t, err)
			err = daemon.UploadConfig(ctx, cv.ID, tarball)
			require.NoError(t, err)

			// create run and wait for it to reach wanted status
			created := daemon.createRun(t, ctx, ws, cv)
			for event := range sub {
				r := event.Payload
				if r.ID == created.ID && r.Status == step.wantStatus {
					// status matches, now check whether reports match as well
					assert.Equal(t, &step.wantResourceReport, event.Payload.Plan.ResourceReport)
					assert.Equal(t, &step.wantOutputReport, r.Plan.OutputReport)
					break
				}
				require.False(t, r.Done(), "run unexpectedly finished with status %s", r.Status)
			}
		})
	}
}
