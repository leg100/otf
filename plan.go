package otf

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/leg100/otf/http/dto"
	httputil "github.com/leg100/otf/http/util"
)

const (
	LocalStateFilename  = "terraform.tfstate"
	PlanFilename        = "plan.out"
	JSONPlanFilename    = "plan.out.json"
	ApplyOutputFilename = "apply.out"
	//List all available plan statuses.
	PlanCanceled    PlanStatus = "canceled"
	PlanCreated     PlanStatus = "created"
	PlanErrored     PlanStatus = "errored"
	PlanFinished    PlanStatus = "finished"
	PlanMFAWaiting  PlanStatus = "mfa_waiting"
	PlanPending     PlanStatus = "pending"
	PlanQueued      PlanStatus = "queued"
	PlanRunning     PlanStatus = "running"
	PlanUnreachable PlanStatus = "unreachable"
)

// Plan represents a Terraform Enterprise plan.
type Plan struct {
	id string
	// Resources is a report of planned resource changes
	*ResourceReport
	// Status is the current status
	status PlanStatus
	// statusTimestamps records timestamps of status transitions
	statusTimestamps []PlanStatusTimestamp
	// run is the parent run
	run *Run
}

func (p *Plan) ID() string         { return p.id }
func (p *Plan) JobID() string      { return p.id }
func (p *Plan) String() string     { return p.id }
func (p *Plan) Status() PlanStatus { return p.status }

// HasChanges determines whether plan has any changes (adds/changes/deletions).
func (p *Plan) HasChanges() bool {
	if p.ResourceReport == nil {
		return false
	}
	if p.Additions > 0 || p.Changes > 0 || p.Destructions > 0 {
		return true
	}
	return false
}

func (p *Plan) GetService(app Application) JobService {
	return app.PlanService()
}

// Do performs a terraform plan
func (p *Plan) Do(env Environment) error {
	if err := p.run.setupEnv(env); err != nil {
		return err
	}
	if err := p.runTerraformPlan(env); err != nil {
		return err
	}
	if err := env.RunCLI("sh", "-c", fmt.Sprintf("terraform show -json %s > %s", PlanFilename, JSONPlanFilename)); err != nil {
		return err
	}
	if err := env.RunFunc(p.run.uploadPlan); err != nil {
		return err
	}
	if err := env.RunFunc(p.run.uploadJSONPlan); err != nil {
		return err
	}
	return nil
}

// Start updates the plan to reflect its plan having started
func (p *Plan) Start(run *Run) error {
	if run.Status() == RunPlanning {
		return ErrJobAlreadyClaimed
	}
	if run.Status() != RunPlanQueued {
		return fmt.Errorf("run cannot be started: invalid status: %s", run.Status())
	}
	run.updateStatus(RunPlanning)
	return nil
}

// Finish updates the run to reflect its plan having finished. An event is
// returned reflecting the run's new status.
func (p *Plan) Finish(opts JobFinishOptions) (*Event, error) {
	if opts.Errored {
		if err := p.run.updateStatus(RunErrored); err != nil {
			return nil, err
		}
		return &Event{Payload: p.run, Type: EventRunErrored}, nil
	}
	if !p.HasChanges() || p.run.Speculative() {
		if err := p.run.updateStatus(RunPlannedAndFinished); err != nil {
			return nil, err
		}
		return &Event{Payload: p.run, Type: EventRunPlannedAndFinished}, nil
	}
	if !p.run.autoApply {
		if err := p.run.updateStatus(RunPlanned); err != nil {
			return nil, err
		}
		return &Event{Payload: p.run, Type: EventRunPlanned}, nil
	}
	if err := p.run.updateStatus(RunApplyQueued); err != nil {
		return nil, err
	}
	return &Event{Type: EventApplyQueued, Payload: p.run}, nil
}

func (p *Plan) StatusTimestamps() []PlanStatusTimestamp { return p.statusTimestamps }

// NewJSONAPIAssembler constructs a PlanJSONAPIAssembler.
func (p *Plan) NewJSONAPIAssembler(req *http.Request, GetPlanLogsRoute string) *PlanJSONAPIAssembler {
	result := &dto.Plan{
		ID:         p.ID(),
		HasChanges: p.HasChanges(),
		LogReadURL: httputil.Absolute(req, fmt.Sprintf(string(GetPlanLogsRoute), p.ID())),
		Status:     string(p.Status()),
	}
	if p.ResourceReport != nil {
		result.ResourceAdditions = p.Additions
		result.ResourceChanges = p.Changes
		result.ResourceDestructions = p.Destructions
	}
	for _, ts := range p.StatusTimestamps() {
		if result.StatusTimestamps == nil {
			result.StatusTimestamps = &dto.PlanStatusTimestamps{}
		}
		switch ts.Status {
		case PlanCanceled:
			result.StatusTimestamps.CanceledAt = &ts.Timestamp
		case PlanErrored:
			result.StatusTimestamps.ErroredAt = &ts.Timestamp
		case PlanFinished:
			result.StatusTimestamps.FinishedAt = &ts.Timestamp
		case PlanQueued:
			result.StatusTimestamps.QueuedAt = &ts.Timestamp
		case PlanRunning:
			result.StatusTimestamps.StartedAt = &ts.Timestamp
		}
	}
	return &PlanJSONAPIAssembler{Plan: result}
}

func (p *Plan) updateStatus(status PlanStatus) {
	p.status = status
	p.statusTimestamps = append(p.statusTimestamps, PlanStatusTimestamp{
		Status:    status,
		Timestamp: CurrentTimestamp(),
	})
}

// runTerraformPlan runs a terraform plan
func (p *Plan) runTerraformPlan(env Environment) error {
	args := []string{
		"plan",
	}
	if p.run.isDestroy {
		args = append(args, "-destroy")
	}
	args = append(args, "-out="+PlanFilename)
	return env.RunCLI("terraform", args...)
}

// PlanJSONAPIAssembler is an intermediatary between an apply and its DTO,
// capable of being assembled into a DTO.
type PlanJSONAPIAssembler struct {
	*dto.Plan
}

// ToJSONAPI implements the jsonapi.Assembler interface.
func (a *PlanJSONAPIAssembler) ToJSONAPI() any {
	return a.Plan
}

// PlanStatus represents a plan state.
type PlanStatus string

type PlanService interface {
	Get(ctx context.Context, id string) (*Plan, error)

	JobService
	ChunkStore
}

type PlanLogStore interface {
	ChunkStore
}

type PlanStatusTimestamp struct {
	Status    PlanStatus
	Timestamp time.Time
}

func newPlan(run *Run) *Plan {
	return &Plan{
		id:  NewID("plan"),
		run: run,
		// new plans always start off in pending state
		status:         PlanPending,
		ResourceReport: &ResourceReport{},
	}
}
