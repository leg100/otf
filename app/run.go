package app

import (
	"fmt"
	"time"

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
		run.UpdateApplyStatus(tfe.ApplyQueued)

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

		// Immediately mark pending/queued runs as cancelled
		switch run.Status {
		case tfe.RunPending, tfe.RunPlanQueued, tfe.RunApplyQueued:
			run.Status = tfe.RunCanceled
			run.StatusTimestamps.CanceledAt = time.Now()
		}

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

func (s RunService) UpdatePlanStatus(id string, status tfe.PlanStatus) (*ots.Run, error) {
	run, err := s.db.Update(id, func(run *ots.Run) error {
		run.UpdatePlanStatus(status)

		return nil
	})
	if err != nil {
		return nil, err
	}
	return run, nil
}

func (s RunService) UpdateApplyStatus(id string, status tfe.ApplyStatus) (*ots.Run, error) {
	run, err := s.db.Update(id, func(run *ots.Run) error {
		run.UpdateApplyStatus(status)

		return nil
	})
	if err != nil {
		return nil, err
	}
	return run, nil
}

func (s RunService) FinishPlan(id string, opts ots.PlanFinishOptions) (*ots.Run, error) {
	run, err := s.db.Update(id, func(run *ots.Run) error {
		run.FinishPlan(opts)

		return nil
	})
	if err != nil {
		return nil, err
	}
	return run, nil
}

func (s RunService) FinishApply(id string, opts ots.ApplyFinishOptions) (*ots.Run, error) {
	run, err := s.db.Update(id, func(run *ots.Run) error {
		run.FinishApply(opts)

		return nil
	})
	if err != nil {
		return nil, err
	}
	return run, nil
}

// GetPlanJSON returns the JSON formatted plan file for the run.
func (s RunService) GetPlanJSON(id string) ([]byte, error) {
	run, err := s.db.Get(ots.RunGetOptions{ID: &id})
	if err != nil {
		return nil, err
	}
	return run.Plan.PlanJSON, nil
}

// GetPlanFile returns the binary plan file for the run.
func (s RunService) GetPlanFile(id string) ([]byte, error) {
	run, err := s.db.Get(ots.RunGetOptions{ID: &id})
	if err != nil {
		return nil, err
	}
	return run.Plan.Plan, nil
}

// GetPlanLogs returns logs from the plan of the run identified by id. The
// options specifies the limit and offset bytes of the logs to retrieve.
func (s RunService) GetPlanLogs(id string, opts ots.PlanLogOptions) ([]byte, error) {
	run, err := s.db.Get(ots.RunGetOptions{PlanID: &id})
	if err != nil {
		return nil, err
	}
	logs := run.Plan.Logs

	// Add start marker
	logs = append([]byte{byte(2)}, logs...)

	// Add end marker
	logs = append(logs, byte(3))

	if opts.Offset > len(logs) {
		return nil, fmt.Errorf("offset cannot be bigger than total logs")
	}

	if opts.Limit > ots.MaxPlanLogsLimit {
		opts.Limit = ots.MaxPlanLogsLimit
	}

	// Ensure specified chunk does not exceed slice length
	if (opts.Offset + opts.Limit) > len(logs) {
		opts.Limit = len(logs) - opts.Offset
	}

	resp := logs[opts.Offset:(opts.Offset + opts.Limit)]

	return resp, nil
}

// GetApplyLogs returns logs from the apply of the run identified by id. The
// options specifies the limit and offset bytes of the logs to retrieve.
func (s RunService) GetApplyLogs(id string, opts ots.ApplyLogOptions) ([]byte, error) {
	run, err := s.db.Get(ots.RunGetOptions{ApplyID: &id})
	if err != nil {
		return nil, err
	}
	logs := run.Apply.Logs

	// Add start marker
	logs = append([]byte{byte(2)}, logs...)

	// Add end marker
	logs = append(logs, byte(3))

	if opts.Offset > len(logs) {
		return nil, fmt.Errorf("offset cannot be bigger than total logs")
	}

	if opts.Limit > ots.MaxApplyLogsLimit {
		opts.Limit = ots.MaxApplyLogsLimit
	}

	// Ensure specified chunk does not exceed slice length
	if (opts.Offset + opts.Limit) > len(logs) {
		opts.Limit = len(logs) - opts.Offset
	}

	resp := logs[opts.Offset:(opts.Offset + opts.Limit)]

	return resp, nil
}

func (s RunService) UploadPlanLogs(id string, logs []byte) error {
	_, err := s.db.Update(id, func(run *ots.Run) error {
		run.Plan.Logs = logs

		return nil
	})
	return err
}
func (s RunService) UploadApplyLogs(id string, logs []byte) error {
	_, err := s.db.Update(id, func(run *ots.Run) error {
		run.Apply.Logs = logs

		return nil
	})
	return err
}
