package agent

import "github.com/leg100/otf/environment"

type fakeJob struct {
	cmd  string
	args []string
}

func (j *fakeJob) Do(e environment.Environment) error {
	return e.RunCLI(j.cmd, j.args...)
}

// discard is a no-op io.WriteCloser
type discard struct{}

func (discard) Write(p []byte) (int, error) {
	return len(p), nil
}

func (discard) Close() error {
	return nil
}
