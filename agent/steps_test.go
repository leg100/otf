package agent

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/leg100/ots"
	"github.com/leg100/ots/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDownloadConfigStep(t *testing.T) {
	cvs := mock.ConfigurationVersionService{
		DownloadFn: func(id string) ([]byte, error) {
			return os.ReadFile("testdata/unpack.tar.gz")
		},
	}
	run := &ots.Run{
		ConfigurationVersion: &ots.ConfigurationVersion{
			ID: "cv-123",
		},
	}
	path := t.TempDir()

	err := DownloadConfigStep(run, cvs).Run(context.Background(), path, nil)
	require.NoError(t, err)

	var got []string
	filepath.Walk(path, func(lpath string, info os.FileInfo, _ error) error {
		lpath, err := filepath.Rel(path, lpath)
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
		"plan.out",
		"terraform.tfstate",
	}, got)
}
