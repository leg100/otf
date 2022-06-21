package otf

// runStateMixin provides common and default behaviour for run state
type runStateMixin struct {
	// whether state is a final state or not
	final bool
	// the job status that corresponds to the run state
	jobStatus JobStatus
}

func (s *runStateMixin) Enqueue() error  { return ErrRunInvalidStateTransition }
func (s *runStateMixin) Cancel() error   { return ErrRunInvalidStateTransition }
func (s *runStateMixin) ApplyRun() error { return ErrRunInvalidStateTransition }
func (s *runStateMixin) Discard() error  { return ErrRunInvalidStateTransition }
func (s *runStateMixin) Start() error    { return ErrRunInvalidStateTransition }
func (s *runStateMixin) Finish(ReportService, JobFinishOptions) error {
	return ErrRunInvalidStateTransition
}

func (s *runStateMixin) Discardable() bool { return false }
func (s *runStateMixin) Confirmable() bool { return false }
func (s *runStateMixin) Cancelable() bool  { return false }
func (s *runStateMixin) Done() bool        { return s.final }
