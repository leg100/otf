package agent

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/leg100/go-tfe"
	"github.com/leg100/ots"
	"github.com/leg100/ots/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type MockRunner struct{}

func (r *MockRunner) Plan(ctx context.Context, path string) ([]byte, error) {
	initOut, err := os.ReadFile("testdata/init.log")
	if err != nil {
		return nil, err
	}
	planOut, err := os.ReadFile("testdata/plan.log")
	if err != nil {
		return nil, err
	}
	return append(initOut, planOut...), nil
}

func TestProcessor(t *testing.T) {
	path := t.TempDir()

	p := &processor{
		ConfigurationVersionService: &mock.ConfigurationVersionService{
			DownloadFn: func(id string) ([]byte, error) {
				return os.ReadFile("testdata/unpack.tar.gz")
			},
		},
		StateVersionService: &mock.StateVersionService{
			CreateFn: func(workspaceID string, opts tfe.StateVersionCreateOptions) (*ots.StateVersion, error) {
				return nil, nil
			},
			CurrentFn: func(workspaceID string) (*ots.StateVersion, error) {
				return &ots.StateVersion{ExternalID: "sv-123"}, nil
			},
			DownloadFn: func(id string) ([]byte, error) {
				return os.ReadFile("testdata/terraform.tfstate")
			},
		},
		RunService: &mock.RunService{
			UploadPlanLogsFn:   func(id string, logs []byte) error { return nil },
			UpdatePlanStatusFn: func(id string, status tfe.PlanStatus) (*ots.Run, error) { return nil, nil },
			FinishPlanFn:       func(id string, opts ots.PlanFinishOptions) (*ots.Run, error) { return nil, nil },
		},
		TerraformRunner: &MockRunner{},
	}

	require.NoError(t, p.Process(context.Background(), &ots.Run{
		Plan: &ots.Plan{
			ExternalID: "plan-123",
		},
		ConfigurationVersion: &ots.ConfigurationVersion{
			ExternalID: "cv-123",
		},
		Workspace: &ots.Workspace{
			ExternalID: "ws-123",
		},
	}, path))

	var got []string
	filepath.Walk(path, func(lpath string, info os.FileInfo, err error) error {
		lpath, err = filepath.Rel(path, lpath)
		require.NoError(t, err)
		got = append(got, lpath)
		return nil
	})
	assert.Equal(t, []string{
		".",
		"dir",
		"dir/file",
		"dir/symlink",
		"file",
		"terraform.tfstate",
	}, got)
}
