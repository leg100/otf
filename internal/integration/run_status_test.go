package integration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/run"
	"github.com/leg100/otf/internal/runstatus"
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
	daemon, org, ctx := setup(t)
	ws, err := daemon.Workspaces.Create(ctx, workspace.CreateOptions{
		Name:         new(t.Name()),
		Organization: &org.Name,
		AutoApply:    new(true),
	})
	require.NoError(t, err)

	// directory for root module
	root := t.TempDir()

	// create a sequence of steps that the test will execute sequentially.
	steps := []struct {
		name               string
		config             string
		wantStatus         runstatus.Status
		wantResourceReport run.Report
		wantOutputReport   run.Report
	}{
		{
			name:       "add resource",
			config:     `resource "random_pet" "cat" { prefix = "mr-" }`,
			wantStatus: runstatus.Applied,
			wantResourceReport: run.Report{
				Additions: 1,
			},
		},
		{
			name: "replace resource",
			config: `resource "random_pet" "cat" { prefix = "sir-" }
`,
			wantStatus: runstatus.Applied,
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
			wantStatus: runstatus.Applied,
			wantOutputReport: run.Report{
				Additions: 1,
			},
		},
		{
			name:       "destroy all",
			wantStatus: runstatus.Applied,
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
			err = daemon.Configs.UploadConfig(ctx, cv.ID, tarball)
			require.NoError(t, err)

			// create run and wait for it to reach wanted status
			created := daemon.createRun(t, ctx, ws, cv, nil)
			updated := daemon.waitRunStatus(t, ctx, created.ID, step.wantStatus)
			// status matches, now check whether reports match as well
			assert.Equal(t, &step.wantResourceReport, updated.Plan.ResourceReport)
			assert.Equal(t, &step.wantOutputReport, updated.Plan.OutputReport)
		})
	}
}
