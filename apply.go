package otf

import (
	"context"
	"fmt"
	"net/http"

	jsonapi "github.com/leg100/otf/http/dto"
	httputil "github.com/leg100/otf/http/util"
)

// Apply represents a terraform apply
type Apply struct {
	id string
	// ResourcesReport is a report of applied resource changes
	*ResourceReport
	// run is the parent run
	run *Run
	// apply is a job
	*phaseMixin
}

func (a *Apply) ID() string      { return a.id }
func (a *Apply) PhaseID() string { return a.id }
func (a *Apply) String() string  { return a.id }

func (a *Apply) Service(app Application) (PhaseService, error) {
	return app.ApplyService(), nil
}

// Do performs a terraform apply
func (a *Apply) Do(env Environment) error {
	if err := a.run.setupEnv(env); err != nil {
		return err
	}
	if err := env.RunFunc(a.run.downloadPlanFile); err != nil {
		return err
	}
	if err := a.runTerraformApply(env); err != nil {
		return err
	}
	if err := env.RunFunc(a.run.uploadState); err != nil {
		return err
	}
	return nil
}

// runTerraformApply runs a terraform apply
func (a *Apply) runTerraformApply(env Environment) error {
	args := []string{"apply"}
	if a.run.isDestroy {
		args = append(args, "-destroy")
	}
	args = append(args, PlanFilename)
	return env.RunCLI("terraform", args...)
}

// ToJSONAPI assembles a JSONAPI DTO.
func (a *Apply) ToJSONAPI(req *http.Request) any {
	dto := &jsonapi.Apply{
		ID:               a.ID(),
		LogReadURL:       httputil.Absolute(req, fmt.Sprintf("applies/%s/logs", a.id)),
		Status:           string(a.Status()),
		StatusTimestamps: &jsonapi.PhaseStatusTimestamps{},
	}
	if a.ResourceReport != nil {
		dto.ResourceAdditions = a.Additions
		dto.ResourceChanges = a.Changes
		dto.ResourceDestructions = a.Destructions
	}
	for _, ts := range a.StatusTimestamps() {
		switch ts.Status {
		case PhaseCanceled:
			dto.StatusTimestamps.CanceledAt = &ts.Timestamp
		case PhaseErrored:
			dto.StatusTimestamps.ErroredAt = &ts.Timestamp
		case PhaseFinished:
			dto.StatusTimestamps.FinishedAt = &ts.Timestamp
		case PhaseQueued:
			dto.StatusTimestamps.QueuedAt = &ts.Timestamp
		case PhaseRunning:
			dto.StatusTimestamps.StartedAt = &ts.Timestamp
		case PhaseUnreachable:
			dto.StatusTimestamps.UnreachableAt = &ts.Timestamp
		}
	}
	return dto
}

// ApplyService allows interaction with Applies
type ApplyService interface {
	Get(ctx context.Context, id string) (*Apply, error)
	PhaseService
	// CreateApplyReport parses the logs from a successful terraform apply and
	// persists a resource report to the database.
	CreateApplyReport(ctx context.Context, applyID string) error
}

func newApply(run *Run) *Apply {
	return &Apply{
		id:         NewID("apply"),
		run:        run,
		phaseMixin: newPhase(),
	}
}
