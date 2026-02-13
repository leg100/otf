package integration

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/leg100/otf/internal"
	runpkg "github.com/leg100/otf/internal/run"
	"github.com/leg100/otf/internal/runstatus"
	"github.com/stretchr/testify/require"
)

// TestIntegration_RunnerParams tests passing of plan parameters.
func TestIntegration_RunnerParams(t *testing.T) {
	integrationTest(t)

	svc, org, ctx := setup(t)

	ws := svc.createWorkspace(t, ctx, org)

	config := fmt.Appendf(nil, `terraform {
  cloud {
    hostname = "%s"
    organization = "%s"

    workspaces {
      name = "%s"
    }
  }
}
resource "random_uuid" "id1" {}
resource "random_uuid" "id2" {}
resource "random_uuid" "id3" {}
`, svc.System.Hostname(), org.Name, ws.Name)

	true_ := true

	applyThing := func(t *testing.T, opt *runpkg.CreateOptions) *runpkg.Report {
		// upload config
		cv := svc.createConfigurationVersion(t, ctx, ws, nil)
		path := t.TempDir()
		err := os.WriteFile(filepath.Join(path, "main.tf"), []byte(config), 0o777)
		require.NoError(t, err)
		tarball, err := internal.Pack(path)
		require.NoError(t, err)
		err = svc.Configs.UploadConfig(ctx, cv.ID, tarball)
		require.NoError(t, err)

		// create run
		opt.AutoApply = &true_
		run := svc.createRun(t, ctx, ws, cv, opt)

		// let it run to completion
		rv := svc.waitRunStatus(t, ctx, run.ID, runstatus.Applied)
		return rv.Plan.ResourceReport
	}

	t.Run("provision the first resource", func(t *testing.T) {
		plan := applyThing(t, &runpkg.CreateOptions{
			TargetAddrs: []string{"random_uuid.id1"},
		})
		require.Equal(t, &runpkg.Report{Additions: 1}, plan)
	})
	t.Run("provision the second resource, replace the first one", func(t *testing.T) {
		plan := applyThing(t, &runpkg.CreateOptions{
			TargetAddrs:  []string{"random_uuid.id1", "random_uuid.id2"},
			ReplaceAddrs: []string{"random_uuid.id1"},
		})
		require.Equal(t, &runpkg.Report{Additions: 2, Destructions: 1}, plan)
	})

	t.Run("delete the first resource", func(t *testing.T) {
		plan := applyThing(t, &runpkg.CreateOptions{
			TargetAddrs: []string{"random_uuid.id1"},
			IsDestroy:   &true_,
		})
		require.Equal(t, &runpkg.Report{Destructions: 1}, plan)
	})
}
