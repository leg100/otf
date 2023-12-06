package run

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/leg100/otf/internal/configversion"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunDownload(t *testing.T) {
	run := &Run{}
	tarball, err := os.ReadFile("./testdata/tarball.tar.gz")
	require.NoError(t, err)
	app := newFakeCLI(run, tarball)

	cmd := app.runDownloadCommand()
	cmd.SetArgs([]string{"run-123"})
	got := bytes.Buffer{}
	cmd.SetOut(&got)

	require.NoError(t, cmd.Execute())
	assert.Regexp(t, `Extracted tarball to: /tmp/run-123-.*`, got.String())
}

type fakeCLIService struct {
	run     *Run
	tarball []byte

	Service
	configversion.ConfigurationVersionService
}

func newFakeCLI(run *Run, tarball []byte) *CLI {
	return &CLI{
		client:  &fakeCLIService{run: run, tarball: tarball},
		configs: &fakeCLIService{run: run, tarball: tarball},
	}
}

func (f *fakeCLIService) GetRun(context.Context, string) (*Run, error) {
	return f.run, nil
}

func (f *fakeCLIService) DownloadConfig(context.Context, string) ([]byte, error) {
	return f.tarball, nil
}
