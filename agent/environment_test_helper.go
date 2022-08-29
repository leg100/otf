package agent

import (
	"bytes"

	"github.com/leg100/otf"
)

type fakeJob struct {
	cmd  string
	args []string
}

func (j *fakeJob) Do(e otf.Environment) error {
	return e.RunCLI(j.cmd, j.args...)
}

type fakeWriteCloser struct {
	*bytes.Buffer
}

func (f *fakeWriteCloser) Write(p []byte) (int, error) {
	return f.Buffer.Write(p)
}

func (*fakeWriteCloser) Close() error {
	return nil
}
