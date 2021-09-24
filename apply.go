package otf

import (
	"fmt"
	"time"

	"gorm.io/gorm"
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
	ID string

	gorm.Model

	Resources

	Status           ApplyStatus
	StatusTimestamps *ApplyStatusTimestamps

	// Logs is the blob ID for the log output from a terraform apply
	LogsBlobID string
}

// ApplyStatusTimestamps holds the timestamps for individual apply statuses.
type ApplyStatusTimestamps struct {
	CanceledAt      *time.Time `json:"canceled-at,omitempty"`
	ErroredAt       *time.Time `json:"errored-at,omitempty"`
	FinishedAt      *time.Time `json:"finished-at,omitempty"`
	ForceCanceledAt *time.Time `json:"force-canceled-at,omitempty"`
	QueuedAt        *time.Time `json:"queued-at,omitempty"`
	StartedAt       *time.Time `json:"started-at,omitempty"`
}

func newApply() *Apply {
	return &Apply{
		ID:               GenerateID("apply"),
		StatusTimestamps: &ApplyStatusTimestamps{},
		LogsBlobID:       NewBlobID(),
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
		return err
	}

	resources, err := parseApplyOutput(string(logs))
	if err != nil {
		return err
	}

	a.Resources = resources

	return nil
}

func (a *Apply) UpdateStatus(status ApplyStatus) {
	a.Status = status
	a.setTimestamp(status)
}

func (a *Apply) setTimestamp(status ApplyStatus) {
	switch status {
	case ApplyCanceled:
		a.StatusTimestamps.CanceledAt = TimeNow()
	case ApplyErrored:
		a.StatusTimestamps.ErroredAt = TimeNow()
	case ApplyFinished:
		a.StatusTimestamps.FinishedAt = TimeNow()
	case ApplyQueued:
		a.StatusTimestamps.QueuedAt = TimeNow()
	case ApplyRunning:
		a.StatusTimestamps.StartedAt = TimeNow()
	}
}
