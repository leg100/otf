package agent

import "github.com/leg100/otf"

type MockSpooler struct {
	queue, cancelations chan otf.Job

	Spooler
}

type MockSpoolerOption func(*MockSpooler)

func WithMockJobs(job ...otf.Job) MockSpoolerOption {
	return func(s *MockSpooler) {
		s.queue = make(chan otf.Job, len(job))
		for _, r := range job {
			s.queue <- r
		}
	}
}

func WithCanceledJobs(job ...otf.Job) MockSpoolerOption {
	return func(s *MockSpooler) {
		s.cancelations = make(chan otf.Job, len(job))
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

func (s *MockSpooler) GetJob() <-chan otf.Job {
	return s.queue
}

func (s *MockSpooler) GetCancelation() <-chan otf.Job {
	return s.queue
}
