package mock

import (
	"github.com/leg100/go-tfe"
	"github.com/leg100/ots"
)

type RunService struct {
	ots.RunService

	GetQueuedFn func(opts tfe.RunListOptions) (*ots.RunList, error)
}

func (s RunService) GetQueued(opts tfe.RunListOptions) (*ots.RunList, error) {
	return s.GetQueuedFn(opts)
}
