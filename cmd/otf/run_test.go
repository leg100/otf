package main

import (
	"bytes"
	"os"
	"testing"

	"github.com/leg100/otf/run"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunDownload(t *testing.T) {
	run := &run.Run{}
	tarball, err := os.ReadFile("./testdata/tarball.tar.gz")
	require.NoError(t, err)
	app := fakeApp(withRun(run), withTarball(tarball))

	cmd := app.runDownloadCommand()
	cmd.SetArgs([]string{"run-123"})
	got := bytes.Buffer{}
	cmd.SetOut(&got)

	require.NoError(t, cmd.Execute())
	assert.Regexp(t, `Extracted tarball to: /tmp/run-123-.*`, got.String())
}
