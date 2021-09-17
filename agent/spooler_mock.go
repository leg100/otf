package agent

import "github.com/leg100/ots"

type MockSpooler struct {
	queue, cancelations chan ots.Job

	Spooler
}

type MockSpoolerOption func(*MockSpooler)

func WithMockJobs(job ...ots.Job) MockSpoolerOption {
	return func(s *MockSpooler) {
		s.queue = make(chan ots.Job, len(job))
		for _, r := range job {
			s.queue <- r
		}
	}
}

func WithCanceledJobs(job ...ots.Job) MockSpoolerOption {
	return func(s *MockSpooler) {
		s.cancelations = make(chan ots.Job, len(job))
		for _, r := range job {
			s.cancelations <- r
		}
	}
}

func NewMockSpooler(opt ...MockSpoolerOption) *MockSpooler {
	spooler := MockSpooler{}
	for _, o := range opt {
		o(&spooler)
	}
	return &spooler
}

func (s *MockSpooler) GetJob() <-chan ots.Job {
	return s.queue
}

func (s *MockSpooler) GetCancelation() <-chan ots.Job {
	return s.queue
}
