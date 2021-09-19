package agent

import (
	"github.com/go-logr/logr"
	"github.com/leg100/otf"
)

type Executor struct {
	logr.Logger

	RunService                  otf.RunService
	ConfigurationVersionService otf.ConfigurationVersionService
	StateVersionService         otf.StateVersionService
}

func (e *Executor) Do(run *otf.Run) error {
	return nil
}

type ExecutorEnvironment struct {
	//lint:ignore U1000 wip
	path string
}
