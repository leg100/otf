package otf

import (
	"context"
	"fmt"
	"net/http"
	"time"

	jsonapi "github.com/leg100/otf/http/dto"
	httputil "github.com/leg100/otf/http/util"
)

const (
	LocalStateFilename = "terraform.tfstate"
	PlanFilename       = "plan.out"
	JSONPlanFilename   = "plan.out.json"
)

// Plan represents a Terraform Enterprise plan.
type Plan struct {
	id string
	// Resources is a report of planned resource changes
	*ResourceReport
	// run is the parent run
	run *Run
	// plan is a job
	*job
}

func (p *Plan) ID() string           { return p.id }
func (p *Plan) String() string       { return p.id }
func (p *Plan) JobStatus() JobStatus { return p.job.status }
func (p *Plan) JobStatusTimestamp(status JobStatus) (time.Time, error) {
	return p.job.StatusTimestamp(status)
}
func (p *Plan) JobStatusTimestamps() []JobStatusTimestamp {
	return p.job.statusTimestamps
}

// HasChanges determines whether plan has any changes (adds/changes/deletions).
func (p *Plan) HasChanges() bool {
	return p.ResourceReport != nil && p.ResourceReport.HasChanges()
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

// Finish updates the run to reflect its plan having finished. An event is
// returned reflecting the run's new status.
func (p *Plan) Finish() (*Event, error) {
	if !p.HasChanges() || p.run.Speculative() {
		if err := p.run.updateStatus(RunPlannedAndFinished); err != nil {
			return nil, err
		}
		return &Event{Payload: p.run, Type: EventRunPlannedAndFinished}, nil
	}
	if !p.run.autoApply {
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
		LogReadURL:       httputil.Absolute(req, fmt.Sprintf("jobs/%s/logs", p.JobID())),
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

type PlanService interface {
	Get(ctx context.Context, id string) (*Plan, error)
}

func newPlan(run *Run) *Plan {
	return &Plan{
		id:  NewID("plan"),
		run: run,
		job: newJob(),
	}
}
