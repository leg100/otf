package agent

import (
	"github.com/go-logr/logr"
	"github.com/leg100/ots"
)

type Executor struct {
	logr.Logger

	RunService                  ots.RunService
	ConfigurationVersionService ots.ConfigurationVersionService
	StateVersionService         ots.StateVersionService
}

func (e *Executor) Do(run *ots.Run) error {
	return nil
}

type ExecutorEnvironment struct {
	//lint:ignore U1000 wip
	path string
}
