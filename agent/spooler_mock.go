package agent

import (
	"context"
)

type mockSpooler struct {
	queue chan Job
}

func newMockSpooler(job ...Job) *mockSpooler {
	queue := make(chan Job, len(job))
	for _, r := range job {
		queue <- r
	}
	return &mockSpooler{queue: queue}
}

func (s *mockSpooler) GetJob() <-chan Job {
	return s.queue
}

func (s *mockSpooler) Start(context.Context) {}
