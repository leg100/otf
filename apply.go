package otf

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	jsonapi "github.com/leg100/otf/http/dto"
	httputil "github.com/leg100/otf/http/util"
)

//List all available apply statuses supported in OTF.
const (
	ApplyCanceled    ApplyStatus = "canceled"
	ApplyErrored     ApplyStatus = "errored"
	ApplyFinished    ApplyStatus = "finished"
	ApplyPending     ApplyStatus = "pending"
	ApplyQueued      ApplyStatus = "queued"
	ApplyRunning     ApplyStatus = "running"
	ApplyUnreachable ApplyStatus = "unreachable"
)

// Apply represents a terraform apply
type Apply struct {
	id    string
	jobID string
	// ResourcesReport is a report of applied resource changes
	*ResourceReport
	// Status is the current status
	status ApplyStatus
	// StatusTimestamps records timestamps of status transitions
	statusTimestamps []ApplyStatusTimestamp
	// run is the parent run
	run *Run
}

func (a *Apply) ID() string          { return a.id }
func (a *Apply) JobID() string       { return a.jobID }
func (a *Apply) String() string      { return a.id }
func (a *Apply) Status() ApplyStatus { return a.status }

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

func (a *Apply) StatusTimestamps() []ApplyStatusTimestamp { return a.statusTimestamps }

func (a *Apply) StatusTimestamp(status ApplyStatus) (time.Time, error) {
	for _, rst := range a.statusTimestamps {
		if rst.Status == status {
			return rst.Timestamp, nil
		}
	}
	return time.Time{}, ErrStatusTimestampNotFound
}

func (a *Apply) updateStatus(status ApplyStatus) {
	a.status = status
	a.statusTimestamps = append(a.statusTimestamps, ApplyStatusTimestamp{
		Status:    status,
		Timestamp: CurrentTimestamp(),
	})
}

// runTerraformApply runs a terraform apply
func (a *Apply) runTerraformApply(env Environment) error {
	cmd := strings.Builder{}
	cmd.WriteString("terraform apply")
	if a.run.isDestroy {
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
		LogReadURL:       httputil.Absolute(req, fmt.Sprintf("jobs/%s/logs", a.JobID())),
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
		case ApplyCanceled:
			dto.StatusTimestamps.CanceledAt = &ts.Timestamp
		case ApplyErrored:
			dto.StatusTimestamps.ErroredAt = &ts.Timestamp
		case ApplyFinished:
			dto.StatusTimestamps.FinishedAt = &ts.Timestamp
		case ApplyQueued:
			dto.StatusTimestamps.QueuedAt = &ts.Timestamp
		case ApplyRunning:
			dto.StatusTimestamps.StartedAt = &ts.Timestamp
		}
	}
	return dto
}

// ApplyStatus represents an apply state.
type ApplyStatus string

// ApplyService allows interaction with Applies
type ApplyService interface {
	Get(ctx context.Context, id string) (*Apply, error)
}

type ApplyStatusTimestamp struct {
	Status    ApplyStatus
	Timestamp time.Time
}

func newApply(run *Run) *Apply {
	return &Apply{
		id:             NewID("apply"),
		jobID:          NewID("job"),
		run:            run,
		status:         ApplyPending,
		ResourceReport: &ResourceReport{},
	}
}
