package agent

type MockJobGetter struct {
	queue chan Job
}

func NewMockJobGetter(job ...Job) *MockJobGetter {
	queue := make(chan Job, len(job))
	for _, r := range job {
		queue <- r
	}
	return &MockJobGetter{queue: queue}
}

func (s *MockJobGetter) GetJob() <-chan Job {
	return s.queue
}
