package otf

import (
	"fmt"
	"time"
)

//List all available apply statuses supported in OTF.
const (
	ApplyCanceled    ApplyStatus = "canceled"
	ApplyCreated     ApplyStatus = "created"
	ApplyErrored     ApplyStatus = "errored"
	ApplyFinished    ApplyStatus = "finished"
	ApplyPending     ApplyStatus = "pending"
	ApplyQueued      ApplyStatus = "queued"
	ApplyRunning     ApplyStatus = "running"
	ApplyUnreachable ApplyStatus = "unreachable"
)

// ApplyStatus represents an apply state.
type ApplyStatus string

type ApplyService interface {
	Get(id string) (*Apply, error)
}

type Apply struct {
	ID string `db:"apply_id"`

	Timestamps

	Resources

	Status           ApplyStatus
	StatusTimestamps TimestampMap

	// Logs is the blob ID for the log output from a terraform apply
	LogsBlobID string

	RunID string
}

func (a *Apply) GetID() string { return a.ID }

func (a *Apply) String() string { return a.ID }

func newApply(runID string) *Apply {
	return &Apply{
		ID:               NewID("apply"),
		Timestamps:       NewTimestamps(),
		StatusTimestamps: make(TimestampMap),
		LogsBlobID:       NewBlobID(),
		RunID:            runID,
	}
}

func (a *Apply) GetLogsBlobID() string {
	return a.LogsBlobID
}

func (a *Apply) Do(run *Run, exe *Executor) error {
	if err := exe.RunFunc(run.downloadPlanFile); err != nil {
		return err
	}

	if err := exe.RunCLI("sh", "-c", fmt.Sprintf("terraform apply -no-color %s | tee %s", PlanFilename, ApplyOutputFilename)); err != nil {
		return err
	}

	if err := exe.RunFunc(run.uploadState); err != nil {
		return err
	}

	return nil
}

// UpdateResources parses the output from terraform apply to determine the
// number and type of resource changes applied and updates the apply object
// accordingly.
func (a *Apply) UpdateResources(bs BlobStore) error {
	logs, err := bs.Get(a.LogsBlobID)
	if err != nil {
		return fmt.Errorf("reading apply logs: %w", err)
	}

	resources, err := parseApplyOutput(string(logs))
	if err != nil {
		return fmt.Errorf("parsing apply output: %w", err)
	}

	a.Resources = resources

	return nil
}

func (a *Apply) UpdateStatus(status ApplyStatus) {
	a.Status = status
	a.setTimestamp(status)
}

func (a *Apply) setTimestamp(status ApplyStatus) {
	a.StatusTimestamps[string(status)] = time.Now()
}
