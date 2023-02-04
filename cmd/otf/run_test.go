package main

import (
	"bytes"
	"os"
	"testing"

	"github.com/leg100/otf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunDownload(t *testing.T) {
	run := otf.NewTestRun(t, otf.TestRunCreateOptions{})
	tarball, err := os.ReadFile("./testdata/tarball.tar.gz")
	require.NoError(t, err)
	app := fakeApp(withFakeRun(run), withFakeTarball(tarball))

	cmd := app.runDownloadCommand()
	cmd.SetArgs([]string{"run-123"})
	got := bytes.Buffer{}
	cmd.SetOut(&got)

	require.NoError(t, cmd.Execute())
	assert.Regexp(t, `Extracted tarball to: /tmp/run-123-.*`, got.String())
}
