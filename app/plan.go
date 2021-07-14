package app

import (
	"fmt"

	"github.com/leg100/go-tfe"
	"github.com/leg100/ots"
)

var _ ots.PlanService = (*PlanService)(nil)

type PlanService struct {
	db ots.RunStore
}

func NewPlanService(db ots.RunStore) *PlanService {
	return &PlanService{
		db: db,
	}
}

func (s PlanService) Get(id string) (*ots.Plan, error) {
	run, err := s.db.Get(ots.RunGetOptions{PlanID: &id})
	if err != nil {
		return nil, err
	}
	return run.Plan, nil
}

func (s PlanService) UpdateStatus(id string, status tfe.PlanStatus) (*ots.Plan, error) {
	run, err := s.db.Update(id, func(run *ots.Run) error {
		run.Plan.UpdateStatus(status)

		return nil
	})
	if err != nil {
		return nil, err
	}
	return run.Plan, nil
}

func (s PlanService) Finish(id string, opts ots.PlanFinishOptions) (*ots.Plan, error) {
	run, err := s.db.Update(id, func(run *ots.Run) error {
		run.Plan.UpdateStatus(tfe.PlanFinished)

		run.Plan.ResourceAdditions = opts.ResourceAdditions
		run.Plan.ResourceChanges = opts.ResourceChanges
		run.Plan.ResourceDestructions = opts.ResourceDestructions

		return nil
	})
	if err != nil {
		return nil, err
	}
	return run.Plan, nil
}

func (s PlanService) GetLogs(id string, opts ots.PlanLogOptions) ([]byte, error) {
	run, err := s.db.Get(ots.RunGetOptions{PlanID: &id})
	if err != nil {
		return nil, err
	}
	logs := run.Plan.Logs

	if opts.Offset > len(logs) {
		return nil, fmt.Errorf("offset too high")
	}
	if opts.Limit > ots.MaxPlanLogsLimit {
		opts.Limit = ots.MaxPlanLogsLimit
	}
	if (opts.Offset + opts.Limit) > len(logs) {
		opts.Limit = len(logs) - opts.Offset
	}

	return logs[opts.Offset:opts.Limit], nil
}

func (s PlanService) UploadLogs(id string, logs []byte) error {
	_, err := s.db.Update(id, func(run *ots.Run) error {
		run.Plan.Logs = logs

		return nil
	})
	return err
}
