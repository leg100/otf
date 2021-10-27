package app

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
)

var _ otf.RunService = (*RunService)(nil)

type RunService struct {
	db otf.RunStore
	es otf.EventService

	logr.Logger

	*otf.RunFactory
}

func NewRunService(db otf.RunStore, logger logr.Logger, wss otf.WorkspaceService, cvs otf.ConfigurationVersionService, es otf.EventService) *RunService {
	return &RunService{
		db:     db,
		es:     es,
		Logger: logger,
		RunFactory: &otf.RunFactory{
			WorkspaceService:            wss,
			ConfigurationVersionService: cvs,
		},
	}
}

// Create constructs and persists a new run object to the db, before scheduling
// the run.
func (s RunService) Create(ctx context.Context, opts otf.RunCreateOptions) (*otf.Run, error) {
	if err := opts.Valid(); err != nil {
		s.Error(err, "creating invalid run")
		return nil, err
	}

	run, err := s.NewRun(opts)
	if err != nil {
		s.Error(err, "constructing new run")
		return nil, err
	}

	_, err = s.db.Create(run)
	if err != nil {
		s.Error(err, "creating run", "id", run.ID)
		return nil, err
	}

	s.V(1).Info("created run", "id", run.ID)

	s.es.Publish(otf.Event{Type: otf.RunCreated, Payload: run})

	return run, nil
}

// Get retrieves a run obj with the given ID from the db.
func (s RunService) Get(id string) (*otf.Run, error) {
	run, err := s.db.Get(otf.RunGetOptions{ID: &id})
	if err != nil {
		s.Error(err, "retrieving run", "id", id)
		return nil, err
	}

	s.V(2).Info("retrieved run", "id", run.ID)

	return run, nil
}

// List retrieves multiple run objs. Use opts to filter and paginate the list.
func (s RunService) List(opts otf.RunListOptions) (*otf.RunList, error) {
	rl, err := s.db.List(opts)
	if err != nil {
		s.Error(err, "listing runs")
		return nil, err
	}

	s.V(2).Info("listed runs", "count", len(rl.Items))

	return rl, nil
}

func (s RunService) Apply(id string, opts otf.RunApplyOptions) error {
	run, err := s.db.Update(id, func(run *otf.Run) error {
		run.UpdateStatus(otf.RunApplyQueued)

		return nil
	})
	if err != nil {
		s.Error(err, "applying run", "id", id)
		return err
	}

	s.V(0).Info("applied run", "id", id)

	s.es.Publish(otf.Event{Type: otf.EventApplyQueued, Payload: run})

	return err
}

func (s RunService) Discard(id string, opts otf.RunDiscardOptions) error {
	run, err := s.db.Update(id, func(run *otf.Run) error {
		return run.Discard()
	})
	if err != nil {
		s.Error(err, "discarding run", "id", id)
		return err
	}

	s.V(0).Info("discarded run", "id", id)

	s.es.Publish(otf.Event{Type: otf.RunCompleted, Payload: run})

	return err
}

// Cancel enqueues a cancel request to cancel a currently queued or active plan
// or apply.
func (s RunService) Cancel(id string, opts otf.RunCancelOptions) error {
	_, err := s.db.Update(id, func(run *otf.Run) error {
		return run.Cancel()
	})
	return err
}

func (s RunService) ForceCancel(id string, opts otf.RunForceCancelOptions) error {
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
		run.UpdateStatus(otf.RunPlanQueued)
		return nil
	})
	if err != nil {
		s.Error(err, "enqueuing plan", "id", id)
		return err
	}

	s.V(0).Info("enqueued plan", "id", id)

	s.es.Publish(otf.Event{Type: otf.EventPlanQueued, Payload: run})

	return err
}

// GetPlanFile returns the plan file for the run.
func (s RunService) GetPlanFile(ctx context.Context, runID string, opts otf.PlanFileOptions) ([]byte, error) {
	getOpts := otf.RunGetOptions{ID: otf.String(runID)}

	switch opts.Format {
	case otf.PlanJSONFormat:
		getOpts.IncludePlanJSON = true
	case otf.PlanBinaryFormat:
		getOpts.IncludePlanFile = true
	default:
		return nil, fmt.Errorf("unknown plan file format specified: %s", opts.Format)
	}

	run, err := s.db.Get(getOpts)
	if err != nil {
		s.Error(err, "retrieving run", "id", runID)
		return nil, err
	}

	s.V(2).Info("retrieved plan file", "id", runID)

	switch opts.Format {
	case otf.PlanJSONFormat:
		return run.Plan.PlanJSON, nil
	case otf.PlanBinaryFormat:
		return run.Plan.PlanFile, nil
	default:
		return nil, fmt.Errorf("unknown plan file format specified: %s", opts.Format)
	}
}

// UploadPlanJSON persists a run's JSON-formatted plan file before parsing it
// and updating the Run's Plan with a summary of planned resource changes. The
// plan file is expected to have been produced using `terraform show -json
// plan_file`.
func (s RunService) UploadPlanJSON(ctx context.Context, id string, plan []byte) error {
	_, err := s.db.Update(id, func(run *otf.Run) error {
		run.Plan.PlanJSON = plan

		return run.Plan.Summarize()
	})
	if err != nil {
		s.Error(err, "uploading plan file in json format", "id", id)
		return err
	}

	s.V(0).Info("uploaded plan file in json format", "id", id)

	return nil
}

// UploadPlanFile persists a run's plan file. The plan file is expected to have
// been produced using `terraform plan`. If the plan file is JSON serialized
// then set json to true.
func (s RunService) UploadPlanFile(ctx context.Context, id string, plan []byte, opts otf.PlanFileOptions) error {
	_, err := s.db.Update(id, func(run *otf.Run) error {
		switch opts.Format {
		case otf.PlanJSONFormat:
			run.Plan.PlanJSON = plan
		case otf.PlanBinaryFormat:
			run.Plan.PlanFile = plan
		default:
			return fmt.Errorf("unknown plan file format specified: %s", opts.Format)
		}

		return nil
	})
	if err != nil {
		s.Error(err, "uploading plan file", "id", id, "format", opts.Format)
		return err
	}

	s.V(0).Info("uploaded plan file", "id", id, "format", opts.Format)

	return nil
}

// Start marks a run phase (plan, apply) as started.
func (s RunService) Start(id string, opts otf.JobStartOptions) (otf.Job, error) {
	run, err := s.db.Update(id, func(run *otf.Run) error {
		return run.Start()
	})
	if err != nil {
		s.Error(err, "starting run")
		return nil, err
	}

	s.V(0).Info("started run", "id", run.ID)

	return run, nil
}

// Finish marks a run phase (plan, apply) as finished.  An event is emitted to
// notify any subscribers of the new run state.
func (s RunService) Finish(id string, opts otf.JobFinishOptions) (otf.Job, error) {
	var event *otf.Event

	run, err := s.db.Update(id, func(run *otf.Run) (err error) {
		event, err = run.Finish(s.bs, s.ls)
		if err != nil {
			return err
		}
		return err
	})
	if err != nil {
		s.Error(err, "finishing run", "id", id)
		return nil, err
	}

	s.V(0).Info("finished run", "id", id)

	s.es.Publish(*event)

	return run, nil
}
