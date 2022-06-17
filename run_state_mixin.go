package otf

import "fmt"

type runStateMixin struct {
	run *Run
}

func (s *runStateMixin) Enqueue() error { return s.err("enqueue") }
func (s *runStateMixin) Cancel() error  { return s.err("cancel") }
func (s *runStateMixin) Apply() error   { return s.err("apply") }
func (s *runStateMixin) Discard() error { return s.err("discard") }
func (s *runStateMixin) Start() error   { return s.err("start") }
func (s *runStateMixin) Finish(RunService) (*ResourceReport, error) {
	return s.err("finish")
}

func (s *runStateMixin) Discardable() bool { return false }
func (s *runStateMixin) Confirmable() bool { return false }
func (s *runStateMixin) Cancelable() bool  { return false }
func (s *runStateMixin) Done() bool        { return false }

func (s *runStateMixin) err(action string) error {
	return fmt.Errorf("invalid state action: %s %s", action, s.run.state)
}
