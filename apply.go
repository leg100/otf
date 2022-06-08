package otf

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	jsonapi "github.com/leg100/otf/http/dto"
	httputil "github.com/leg100/otf/http/util"
)

// Apply represents a terraform apply
type Apply struct {
	id string
	// ResourcesReport is a report of applied resource changes
	*ResourceReport
	// A plan is a job
	*job
}

func (a *Apply) ID() string     { return a.id }
func (a *Apply) String() string { return a.id }

// Do sets up and runs terraform apply.
func (a *Apply) Do(env Environment) error {
	if err := a.setup(env); err != nil {
		return err
	}
	if err := env.RunFunc(a.downloadPlanFile); err != nil {
		return err
	}
	if err := a.apply(env); err != nil {
		return err
	}
	if err := env.RunFunc(a.uploadState); err != nil {
		return err
	}
	return nil
}

// Start updates the run to reflect its apply having started
func (a *Apply) Start() error {
	if a.status == JobRunning {
		return ErrJobAlreadyClaimed
	}
	if a.status != JobQueued {
		return fmt.Errorf("run cannot be started: invalid status: %s", a.status)
	}
	a.UpdateStatus(JobRunning)
	return nil
}

// Finish updates the run to reflect its apply having finished. An event is
// returned reflecting the run's new status.
func (a *Apply) Finish() error {
	a.UpdateStatus(JobFinished)
	return nil
}

// TODO: return a command string instead and have Do() execute it - this'll make
// it more suitable for unit testing.
//
// apply executes terraform apply
func (a *Apply) apply(env Environment) error {
	cmd := strings.Builder{}
	cmd.WriteString("terraform apply")
	if a.isDestroy {
		cmd.WriteString(" -destroy")
	}
	cmd.WriteRune(' ')
	cmd.WriteString(PlanFilename)
	cmd.WriteString(" | tee ")
	cmd.WriteString(ApplyOutputFilename)
	return env.RunCLI("sh", "-c", cmd.String())
}

// ToJSONAPI assembles a JSONAPI DTO.
func (a *Apply) ToJSONAPI(req *http.Request) any {
	dto := &jsonapi.Apply{
		ID:               a.ID(),
		LogReadURL:       httputil.Absolute(req, fmt.Sprintf("applies/%s/logs", a.ID())),
		Status:           string(a.Status()),
		StatusTimestamps: &jsonapi.ApplyStatusTimestamps{},
	}
	if a.ResourceReport != nil {
		dto.ResourceAdditions = a.Additions
		dto.ResourceChanges = a.Changes
		dto.ResourceDestructions = a.Destructions
	}
	for _, ts := range a.StatusTimestamps() {
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

// ApplyService allows interaction with Applies
type ApplyService interface {
	Get(ctx context.Context, id string) (*Apply, error)
}

func newApply(run *Run) *Apply {
	return &Apply{
		id:             NewID("apply"),
		job:            newJob(),
		ResourceReport: &ResourceReport{},
	}
}
