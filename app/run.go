package app

import (
	"fmt"

	"github.com/leg100/go-tfe"
	"github.com/leg100/ots"
)

var _ ots.RunService = (*RunService)(nil)

type RunService struct {
	db ots.RunStore
	bs ots.BlobStore
	es ots.EventService

	*ots.RunFactory
}

func NewRunService(db ots.RunStore, wss ots.WorkspaceService, cvs ots.ConfigurationVersionService, bs ots.BlobStore, es ots.EventService) *RunService {
	return &RunService{
		bs: bs,
		db: db,
		es: es,
		RunFactory: &ots.RunFactory{
			WorkspaceService:            wss,
			ConfigurationVersionService: cvs,
		},
	}
}

// Create constructs and persists a new run object to the db, before scheduling
// the run.
func (s RunService) Create(opts *tfe.RunCreateOptions) (*ots.Run, error) {
	run, err := s.NewRun(opts)
	if err != nil {
		return nil, err
	}

	run, err = s.db.Create(run)
	if err != nil {
		return nil, err
	}

	s.es.Publish(ots.Event{Type: ots.RunCreated, Payload: run})

	return run, nil
}

// Get retrieves a run obj with the given ID from the db.
func (s RunService) Get(id string) (*ots.Run, error) {
	return s.db.Get(ots.RunGetOptions{ID: &id})
}

// List retrieves multiple run objs. Use opts to filter and paginate the list.
func (s RunService) List(opts ots.RunListOptions) (*ots.RunList, error) {
	return s.db.List(opts)
}

func (s RunService) Apply(id string, opts *tfe.RunApplyOptions) error {
	run, err := s.db.Update(id, func(run *ots.Run) error {
		run.UpdateApplyStatus(tfe.ApplyQueued)

		return nil
	})
	if err != nil {
		return err
	}

	s.es.Publish(ots.Event{Type: ots.ApplyQueued, Payload: run})

	return err
}

func (s RunService) Discard(id string, opts *tfe.RunDiscardOptions) error {
	run, err := s.db.Update(id, func(run *ots.Run) error {
		return run.Discard()
	})
	if err != nil {
		return err
	}

	s.es.Publish(ots.Event{Type: ots.RunCompleted, Payload: run})

	return err
}

// Cancel enqueues a cancel request to cancel a currently queued or active plan
// or apply.
func (s RunService) Cancel(id string, opts *tfe.RunCancelOptions) error {
	_, err := s.db.Update(id, func(run *ots.Run) error {
		if err := run.IssueCancel(); err != nil {
			return err
		}

		// Immediately mark pending/queued runs as cancelled
		switch run.Status {
		case tfe.RunPending, tfe.RunPlanQueued, tfe.RunApplyQueued:
			run.Status = tfe.RunCanceled
			run.StatusTimestamps.CanceledAt = ots.TimeNow()
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

func (s RunService) EnqueuePlan(id string) error {
	run, err := s.db.Update(id, func(run *ots.Run) error {
		return run.UpdateStatus(tfe.RunPlanQueued)
	})
	if err != nil {
		return err
	}

	s.es.Publish(ots.Event{Type: ots.PlanQueued, Payload: run})

	return err
}

func (s RunService) UpdateStatus(id string, status tfe.RunStatus) (*ots.Run, error) {
	return s.db.Update(id, func(run *ots.Run) error {
		return run.UpdateStatus(status)
	})
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

// UploadPlan persists a run's plan file. The plan file is expected to have been
// produced using `terraform plan`. If the plan file is JSON serialized then set
// json to true.
func (s RunService) UploadPlan(id string, plan []byte, json bool) error {
	blobID, err := s.bs.Put(plan)
	if err != nil {
		return err
	}

	_, err = s.db.Update(id, func(run *ots.Run) error {
		if json {
			run.Plan.PlanJSONBlobID = blobID
		} else {
			run.Plan.PlanFileBlobID = blobID
		}

		return nil
	})
	return err
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

	s.es.Publish(ots.Event{Type: ots.RunCompleted, Payload: run})

	return run, nil
}

// GetPlanJSON returns the JSON formatted plan file for the run.
func (s RunService) GetPlanJSON(id string) ([]byte, error) {
	run, err := s.db.Get(ots.RunGetOptions{ID: &id})
	if err != nil {
		return nil, err
	}
	return s.bs.Get(run.Plan.PlanJSONBlobID)
}

// GetPlanFile returns the binary plan file for the run.
func (s RunService) GetPlanFile(id string) ([]byte, error) {
	run, err := s.db.Get(ots.RunGetOptions{ID: &id})
	if err != nil {
		return nil, err
	}
	return s.bs.Get(run.Plan.PlanFileBlobID)
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
