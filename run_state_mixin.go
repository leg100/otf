package otf

// runStateMixin provides default behaviour for run state
type runStateMixin struct{}

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
func (s *runStateMixin) Done() bool        { return false }
