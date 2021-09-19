package agent

import "github.com/leg100/otf"

type MockJobGetter struct {
	queue chan otf.Job
}

func NewMockJobGetter(job ...otf.Job) *MockJobGetter {
	queue := make(chan otf.Job, len(job))
	for _, r := range job {
		queue <- r
	}
	return &MockJobGetter{queue: queue}
}

func (s *MockJobGetter) GetJob() <-chan otf.Job {
	return s.queue
}
