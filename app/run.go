package app

import (
	"github.com/leg100/go-tfe"
	"github.com/leg100/ots"
)

var _ ots.RunService = (*RunService)(nil)

type RunService struct {
	db ots.RunStore

	*ots.RunFactory
}

func NewRunService(db ots.RunStore, wss ots.WorkspaceService, cvs ots.ConfigurationVersionService) *RunService {
	return &RunService{
		db: db,
		RunFactory: &ots.RunFactory{
			WorkspaceService:            wss,
			ConfigurationVersionService: cvs,
		},
	}
}

func (s RunService) Create(opts *tfe.RunCreateOptions) (*ots.Run, error) {
	run, err := s.NewRun(opts)
	if err != nil {
		return nil, err
	}

	return s.db.Create(run)
}

func (s RunService) Get(id string) (*ots.Run, error) {
	return s.db.Get(ots.RunGetOptions{ID: &id})
}

// GetQueuedRuns retrieves a list of runs with current status of RunPlanQueued
// or RunApplyQueued.
func (s RunService) List(workspaceID string, opts tfe.RunListOptions) (*ots.RunList, error) {
	dopts := ots.RunListOptions{
		ListOptions: opts.ListOptions,
		WorkspaceID: &workspaceID,
	}

	return s.db.List(dopts)
}

// GetQueuedRuns retrieves a list of runs with current status of RunPlanQueued
// or RunApplyQueued.
func (s RunService) GetQueued(opts tfe.RunListOptions) (*ots.RunList, error) {
	dopts := ots.RunListOptions{
		ListOptions: opts.ListOptions,
		Statuses:    []tfe.RunStatus{tfe.RunPlanQueued, tfe.RunApplyQueued},
	}

	return s.db.List(dopts)
}

func (s RunService) Apply(id string, opts *tfe.RunApplyOptions) error {
	_, err := s.db.Update(id, func(run *ots.Run) error {
		run.QueueApply()

		return nil
	})
	return err
}

func (s RunService) Discard(id string, opts *tfe.RunDiscardOptions) error {
	_, err := s.db.Update(id, func(run *ots.Run) error {
		return run.Discard()
	})
	return err
}

// CancelRun enqueues a cancel request to cancel a currently queued or active
// plan or apply.
func (s RunService) Cancel(id string, opts *tfe.RunCancelOptions) error {
	_, err := s.db.Update(id, func(run *ots.Run) error {
		if err := run.IssueCancel(); err != nil {
			return err
		}

		// TODO: remove plan/apply from queue if queued

		// TODO: send INT signal to terraform process

		return nil
	})
	return err
}

func (s RunService) ForceCancel(id string, opts *tfe.RunForceCancelOptions) error {
	_, err := s.db.Update(id, func(run *ots.Run) error {
		if err := run.ForceCancel(); err != nil {
			return err
		}

		// TODO: send KILL signal to running terraform process

		// TODO: unlock workspace

		return nil
	})

	return err
}
