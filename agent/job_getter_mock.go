package agent

import "github.com/leg100/ots"

type MockJobGetter struct {
	queue chan ots.Job
}

func NewMockJobGetter(job ...ots.Job) *MockJobGetter {
	queue := make(chan ots.Job, len(job))
	for _, r := range job {
		queue <- r
	}
	return &MockJobGetter{queue: queue}
}

func (s *MockJobGetter) GetJob() <-chan ots.Job {
	return s.queue
}
