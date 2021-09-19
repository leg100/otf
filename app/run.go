package app

import (
	"context"
	"fmt"

	"github.com/leg100/go-tfe"
	"github.com/leg100/otf"
)

var _ otf.RunService = (*RunService)(nil)

type RunService struct {
	db otf.RunStore
	bs otf.BlobStore
	es otf.EventService

	*otf.RunFactory
}

func NewRunService(db otf.RunStore, wss otf.WorkspaceService, cvs otf.ConfigurationVersionService, bs otf.BlobStore, es otf.EventService) *RunService {
	return &RunService{
		bs: bs,
		db: db,
		es: es,
		RunFactory: &otf.RunFactory{
			WorkspaceService:            wss,
			ConfigurationVersionService: cvs,
		},
	}
}

// Create constructs and persists a new run object to the db, before scheduling
// the run.
func (s RunService) Create(opts *tfe.RunCreateOptions) (*otf.Run, error) {
	run, err := s.NewRun(opts)
	if err != nil {
		return nil, err
	}

	run, err = s.db.Create(run)
	if err != nil {
		return nil, err
	}

	s.es.Publish(otf.Event{Type: otf.RunCreated, Payload: run})

	return run, nil
}

// Get retrieves a run obj with the given ID from the db.
func (s RunService) Get(id string) (*otf.Run, error) {
	return s.db.Get(otf.RunGetOptions{ID: &id})
}

// List retrieves multiple run objs. Use opts to filter and paginate the list.
func (s RunService) List(opts otf.RunListOptions) (*otf.RunList, error) {
	return s.db.List(opts)
}

func (s RunService) Apply(id string, opts *tfe.RunApplyOptions) error {
	run, err := s.db.Update(id, func(run *otf.Run) error {
		run.UpdateStatus(tfe.RunApplyQueued)

		return nil
	})
	if err != nil {
		return err
	}

	s.es.Publish(otf.Event{Type: otf.ApplyQueued, Payload: run})

	return err
}

func (s RunService) Discard(id string, opts *tfe.RunDiscardOptions) error {
	run, err := s.db.Update(id, func(run *otf.Run) error {
		return run.Discard()
	})
	if err != nil {
		return err
	}

	s.es.Publish(otf.Event{Type: otf.RunCompleted, Payload: run})

	return err
}

// Cancel enqueues a cancel request to cancel a currently queued or active plan
// or apply.
func (s RunService) Cancel(id string, opts *tfe.RunCancelOptions) error {
	_, err := s.db.Update(id, func(run *otf.Run) error {
		return run.Cancel()
	})
	return err
}

func (s RunService) ForceCancel(id string, opts *tfe.RunForceCancelOptions) error {
	_, err := s.db.Update(id, func(run *otf.Run) error {
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
	run, err := s.db.Update(id, func(run *otf.Run) error {
		run.UpdateStatus(tfe.RunPlanQueued)
		return nil
	})
	if err != nil {
		return err
	}

	s.es.Publish(otf.Event{Type: otf.PlanQueued, Payload: run})

	return err
}

// UploadPlan persists a run's plan file. The plan file is expected to have been
// produced using `terraform plan`. If the plan file is JSON serialized then set
// json to true.
func (s RunService) UploadPlan(ctx context.Context, id string, plan []byte, opts tfe.RunUploadPlanOptions) error {
	run, err := s.db.Get(otf.RunGetOptions{ID: &id})
	if err != nil {
		return err
	}

	var bid string // Blob ID

	switch opts.Format {
	case tfe.PlanJSONFormat:
		bid = run.Plan.PlanJSONBlobID
	case tfe.PlanBinaryFormat:
		bid = run.Plan.PlanFileBlobID
	default:
		return fmt.Errorf("unknown plan file format specified: %s", opts.Format)
	}

	return s.bs.Put(bid, plan)
}

// Start marks a run phase (plan, apply) as started.
func (s RunService) Start(id string, opts otf.JobStartOptions) (otf.Job, error) {
	return s.db.Update(id, func(run *otf.Run) error {
		return run.Start()
	})
}

// Finish marks a run phase (plan, apply) as finished.  An event is emitted to
// notify any subscribers of the new run state.
func (s RunService) Finish(id string, opts otf.JobFinishOptions) (otf.Job, error) {
	var event *otf.Event

	run, err := s.db.Update(id, func(run *otf.Run) (err error) {
		event, err = run.Finish(s.bs)
		if err != nil {
			return err
		}
		return err
	})
	if err != nil {
		return nil, err
	}

	s.es.Publish(*event)

	return run, nil
}

// GetPlanJSON returns the JSON formatted plan file for the run.
func (s RunService) GetPlanJSON(id string) ([]byte, error) {
	run, err := s.db.Get(otf.RunGetOptions{ID: &id})
	if err != nil {
		return nil, err
	}
	return s.bs.Get(run.Plan.PlanJSONBlobID)
}

// GetPlanFile returns the binary plan file for the run.
func (s RunService) GetPlanFile(id string) ([]byte, error) {
	run, err := s.db.Get(otf.RunGetOptions{ID: &id})
	if err != nil {
		return nil, err
	}
	return s.bs.Get(run.Plan.PlanFileBlobID)
}

// GetPlanLogs returns logs from the plan of the run identified by id. The
// options specifies the limit and offset bytes of the logs to retrieve.
func (s RunService) GetPlanLogs(id string, opts otf.GetChunkOptions) ([]byte, error) {
	run, err := s.db.Get(otf.RunGetOptions{PlanID: &id})
	if err != nil {
		return nil, err
	}
	return s.bs.GetChunk(run.Plan.LogsBlobID, opts)
}

// GetApplyLogs returns logs from the apply of the run identified by id. The
// options specifies the limit and offset bytes of the logs to retrieve.
func (s RunService) GetApplyLogs(id string, opts otf.GetChunkOptions) ([]byte, error) {
	run, err := s.db.Get(otf.RunGetOptions{ApplyID: &id})
	if err != nil {
		return nil, err
	}
	return s.bs.GetChunk(run.Apply.LogsBlobID, opts)
}

// UploadLogs writes a chunk of logs for a run.
func (s RunService) UploadLogs(id string, logs []byte, opts otf.PutChunkOptions) error {
	run, err := s.db.Get(otf.RunGetOptions{ID: &id})
	if err != nil {
		return err
	}

	active, err := run.ActivePhase()
	if err != nil {
		return fmt.Errorf("attempted to upload logs to an inactive run: %w", err)
	}

	return s.bs.PutChunk(active.GetLogsBlobID(), logs, opts)
}
