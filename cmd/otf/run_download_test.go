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
	factory := &fakeClientFactory{run: run, tarball: tarball}
	require.NoError(t, err)

	cmd := RunDownloadCommand(factory)
	cmd.SetArgs([]string{"run-123"})
	got := bytes.Buffer{}
	cmd.SetOut(&got)

	require.NoError(t, cmd.Execute())
	assert.Regexp(t, `Extracted tarball to: /tmp/run-123-.*`, got.String())
}
