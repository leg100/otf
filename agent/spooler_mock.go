package agent

import (
	"context"

	"github.com/leg100/ots"
)

type mockSpooler struct {
	queue chan *ots.Run
}

func newMockSpooler(run ...*ots.Run) *mockSpooler {
	queue := make(chan *ots.Run, len(run))
	for _, r := range run {
		queue <- r
	}
	return &mockSpooler{queue: queue}
}

func (s *mockSpooler) GetJob() <-chan *ots.Run {
	return s.queue
}

func (s *mockSpooler) Start(context.Context) {}
