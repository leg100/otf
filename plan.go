package otf

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	jsonapi "github.com/leg100/otf/http/dto"
	httputil "github.com/leg100/otf/http/util"
)

const (
	LocalStateFilename  = "terraform.tfstate"
	PlanFilename        = "plan.out"
	JSONPlanFilename    = "plan.out.json"
	ApplyOutputFilename = "apply.out"
)

// Plan represents a "terraform plan"
type Plan struct {
	id string
	// Resources is a report of planned resource changes
	*ResourceReport
	// A plan is a job
	*job
	// run is the parent run
	run *Run
}

func (p *Plan) ID() string     { return p.id }
func (p *Plan) String() string { return p.id }

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

// Do performs a terraform plan
func (p *Plan) Do(env Environment) error {
	if err := p.setup(env); err != nil {
		return err
	}
	if err := p.plan(env); err != nil {
		return err
	}
	if err := env.RunCLI("sh", "-c", fmt.Sprintf("terraform show -json %s > %s", PlanFilename, JSONPlanFilename)); err != nil {
		return err
	}
	if err := env.RunFunc(p.uploadPlan); err != nil {
		return err
	}
	if err := env.RunFunc(p.uploadJSONPlan); err != nil {
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

// ToJSONAPI assembles a JSON-API DTO.
func (p *Plan) ToJSONAPI(req *http.Request) any {
	dto := &jsonapi.Plan{
		ID:               p.ID(),
		HasChanges:       p.HasChanges(),
		LogReadURL:       httputil.Absolute(req, fmt.Sprintf("plans/%s/logs", p.ID())),
		Status:           string(p.Status()),
		StatusTimestamps: &jsonapi.PlanStatusTimestamps{},
	}
	if p.ResourceReport != nil {
		dto.ResourceAdditions = p.Additions
		dto.ResourceChanges = p.Changes
		dto.ResourceDestructions = p.Destructions
	}
	for _, ts := range p.StatusTimestamps() {
		switch ts.Status {
		case JobCanceled:
			dto.StatusTimestamps.CanceledAt = &ts.Timestamp
		case JobErrored:
			dto.StatusTimestamps.ErroredAt = &ts.Timestamp
		case JobFinished:
			dto.StatusTimestamps.FinishedAt = &ts.Timestamp
		case JobQueued:
			dto.StatusTimestamps.QueuedAt = &ts.Timestamp
		case JobRunning:
			dto.StatusTimestamps.StartedAt = &ts.Timestamp
		}
	}
	return dto
}

// TODO: return a command string instead and have Do() execute it - this'll make
// it more suitable for unit testing.
//
// plan executes terraform plan
func (p *Plan) plan(env Environment) error {
	args := []string{
		"plan",
	}
	if p.isDestroy {
		args = append(args, "-destroy")
	}
	args = append(args, "-out="+PlanFilename)
	return env.RunCLI("terraform", args...)
}

func (p *Plan) uploadPlan(ctx context.Context, env Environment) error {
	file, err := os.ReadFile(filepath.Join(env.Path(), PlanFilename))
	if err != nil {
		return err
	}

	if err := env.RunService().UploadPlanFile(ctx, p.ID(), file, PlanFormatBinary); err != nil {
		return fmt.Errorf("unable to upload plan: %w", err)
	}

	return nil
}

func (p *Plan) uploadJSONPlan(ctx context.Context, env Environment) error {
	jsonFile, err := os.ReadFile(filepath.Join(env.Path(), JSONPlanFilename))
	if err != nil {
		return err
	}
	if err := env.RunService().UploadPlanFile(ctx, p.ID(), jsonFile, PlanFormatJSON); err != nil {
		return fmt.Errorf("unable to upload JSON plan: %w", err)
	}
	return nil
}

type PlanService interface {
	Get(ctx context.Context, id string) (*Plan, error)
}

func newPlan(run *Run) *Plan {
	return &Plan{
		id:             NewID("plan"),
		job:            newJob(),
		ResourceReport: &ResourceReport{},
	}
}
